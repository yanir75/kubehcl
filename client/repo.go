package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"kubehcl.sh/kubehcl/internal/configs"
	"kubehcl.sh/kubehcl/internal/decode"
	"kubehcl.sh/kubehcl/internal/logging"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func parseRepoAddArgs(args []string) (string, string, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	if len(args) != 2 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Required arguments are :[name, URL]",
			Detail:   fmt.Sprintf("Got %s", args),
		})
		return "", "", diags
	}
	return args[0], args[1], diags

}

func generateBlockFromRepo(opts *settings.RepoAddOptions) *hclwrite.Block {
	block := hclwrite.NewBlock("repo", []string{opts.Name})
	valueOf := reflect.ValueOf(*opts)
	typeOf := reflect.TypeOf(*opts)

	for i := 0; i < valueOf.NumField(); i++ {

		val := valueOf.Field(i).Interface()

		switch tt := val.(type) {
		case int64:
			block.Body().SetAttributeValue(typeOf.Field(i).Name, cty.NumberIntVal(tt))
		case string:
			block.Body().SetAttributeValue(typeOf.Field(i).Name, cty.StringVal(tt))
		case bool:
			block.Body().SetAttributeValue(typeOf.Field(i).Name, cty.BoolVal(tt))
		case time.Duration:
			block.Body().SetAttributeValue(typeOf.Field(i).Name, cty.NumberIntVal(int64(tt.Seconds())))
		default:
			panic("shouldn't get here")
		}

	}
	return block
}

func generateBlockFromValue(value *decode.DecodedRepo) *hclwrite.Block {
	block := hclwrite.NewBlock("repo", []string{value.Name})
	valueOf := reflect.ValueOf(*value)
	typeOf := reflect.TypeOf(*value)

	for i := 0; i < valueOf.NumField(); i++ {

		val := valueOf.Field(i).Interface()

		switch tt := val.(type) {
		case int64:
			block.Body().SetAttributeValue(typeOf.Field(i).Name, cty.NumberIntVal(tt))
		case string:
			block.Body().SetAttributeValue(typeOf.Field(i).Name, cty.StringVal(tt))
		case bool:
			block.Body().SetAttributeValue(typeOf.Field(i).Name, cty.BoolVal(tt))
		case time.Duration:
			block.Body().SetAttributeValue(typeOf.Field(i).Name, cty.NumberIntVal(int64(tt.Seconds())))
		default:
			logger := logging.KubeLogger
			logger.Debug(fmt.Sprintf("Unused field %s", typeOf.Field(i).Name))
		}

	}
	return block
}

func AddRepo(opts *settings.RepoAddOptions, envSettings *settings.EnvSettings, viewDef *view.ViewArgs, args []string) {
	name, u, diags := parseRepoAddArgs(args)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return
	}

	parsedUrl, err := url.Parse(u)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't parse url",
			Detail:   fmt.Sprintf("Url %s can't be parsed, err: %s", u, err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return
	}
	if strings.Contains(u, "://") {
		opts.Url = parsedUrl.Host + parsedUrl.Path
	} else {
		opts.Url = u
	}
	opts.Name = name
	if parsedUrl.Scheme == "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "No protocol",
			Detail:   fmt.Sprintf("Url %s doesn't contain protocol, please add protocol like https:// or oci://", u),
		})
		v.DiagPrinter(diags, viewDef)
		return
	}
	opts.Protocol = parsedUrl.Scheme

	if parsedUrl.Scheme == "oci" {
		AddRepoOci(opts, envSettings, viewDef)
	} else {
		AddRepoHttp(opts, envSettings, viewDef)
	}

}

func doRequest(opts *settings.RepoAddOptions, path string, httpClient *http.Client, fullUrl string) (*bytes.Buffer, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var err error
	var req *http.Request

	if fullUrl == "" {
		req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s://%s/%s", opts.Protocol, opts.Url, path), nil)
	} else {
		req, err = http.NewRequest(http.MethodGet, fullUrl, nil)
	}

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't create request",
			Detail:   fmt.Sprintf("Request couldn't be created err: %s", err.Error()),
		})
		return nil, diags
	}

	req.Header.Set("User-Agent", "kubehcl")

	if opts.Username != "" && opts.Password != "" {
		req.SetBasicAuth(opts.Username, opts.Password)
	}

	resp, err := httpClient.Do(req)

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Request failed",
			Detail:   fmt.Sprintf("Request couldn't be created err: %s", err.Error()),
		})
		return nil, diags
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Request failed",
			Detail:   fmt.Sprintf("Status code is %d", http.StatusOK),
		})
		return nil, diags
	}

	buf := bytes.NewBuffer(nil)

	_, err = io.Copy(buf, resp.Body)

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't copy body to buff",
			Detail:   fmt.Sprint(err.Error()),
		})
		return nil, diags
	}

	return buf, diags
}

func AddRepoHttp(opts *settings.RepoAddOptions, envSettings *settings.EnvSettings, viewDef *view.ViewArgs) {
	repos, diags := configs.DecodeRepos(envSettings.RepositoryConfig)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return
	}

	if val, ok := repos[opts.Name]; ok {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Repository already exists",
			Detail:   fmt.Sprintf("Repository %s already exists", val.Name),
			Subject:  &val.DeclRange,
		})
		v.DiagPrinter(diags, viewDef)
		return
	}

	opts.RepoCache = envSettings.RepositoryCache
	opts.RepoFile = envSettings.RepositoryConfig

	httpClient, diags := newHttpClient(opts)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return
	}
	_, diags = doRequest(opts, "index.yaml", httpClient, "")
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return
	}

	f := hclwrite.NewEmptyFile()
	body := f.Body()
	body.AppendBlock(generateBlockFromRepo(opts))

	repoCacheFile, err := os.OpenFile(envSettings.RepositoryConfig, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't create file",
			Detail:   fmt.Sprint(err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return
	}

	_, err = f.WriteTo(repoCacheFile)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't write to file",
			Detail:   fmt.Sprint(err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return
	}

	_, err = repoCacheFile.WriteString("\n")

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't write to file",
			Detail:   fmt.Sprint(err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return
	}
}

func AddRepoOci(opts *settings.RepoAddOptions, envSettings *settings.EnvSettings, viewDef *view.ViewArgs) {
	repos, diags := configs.DecodeRepos(envSettings.RepositoryConfig)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return
	}

	if val, ok := repos[opts.Name]; ok {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Repository already exists",
			Detail:   fmt.Sprintf("Repository %s already exists", val.Name),
			Subject:  &val.DeclRange,
		})
		v.DiagPrinter(diags, viewDef)
		return
	}

	opts.RepoCache = envSettings.RepositoryCache
	opts.RepoFile = envSettings.RepositoryConfig
	repo, err := remote.NewRepository(opts.Url)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't load CA certificate",
			Detail:   fmt.Sprint(err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return
	}

	repo.Client, diags = newAuthClient(opts, repo.Reference.Registry)

	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return
	}

	repo.PlainHTTP = opts.PlainHttp
	repo.TagListPageSize = 1
	err = repo.Tags(context.Background(), "", func(tags []string) error {
		// fmt.Printf("%s",tags)
		return nil
	})
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't authenticate",
			Detail:   fmt.Sprint(err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return
	}

	f := hclwrite.NewEmptyFile()
	body := f.Body()
	body.AppendBlock(generateBlockFromRepo(opts))

	repoCacheFile, err := os.OpenFile(envSettings.RepositoryConfig, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't create file",
			Detail:   fmt.Sprint(err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return
	}

	_, err = f.WriteTo(repoCacheFile)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't write to file",
			Detail:   fmt.Sprint(err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return
	}

	_, err = repoCacheFile.WriteString("\n")

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't write to file",
			Detail:   fmt.Sprint(err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return
	}
}

func newHttpClient(opts *settings.RepoAddOptions) (*http.Client, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	cfg := &tls.Config{
		InsecureSkipVerify: opts.InsecureSkipTLSverify,
	}
	if opts.KeyFile != "" && opts.CertFile != "" {
		cert, err := tls.LoadX509KeyPair(opts.CertFile, opts.KeyFile)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Couldn't load certificate",
				Detail:   fmt.Sprint(err.Error()),
			})
			return nil, diags
		}
		cfg.Certificates = append(cfg.Certificates, cert)
	}
	if opts.CaFile != "" {
		caCert, err := os.ReadFile(opts.CaFile)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Couldn't load CA certificate",
				Detail:   fmt.Sprint(err.Error()),
			})
			return nil, diags
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		cfg.RootCAs = caCertPool
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: cfg,
		},
	}

	return httpClient, diags
}

func newAuthClient(opts *settings.RepoAddOptions, regName string) (*auth.Client, hcl.Diagnostics) {
	httpClient, diags := newHttpClient(opts)
	if diags.HasErrors() {
		return nil, diags
	}

	authClient := &auth.Client{
		Client: httpClient,
		Cache:  auth.NewCache(),
		Credential: auth.StaticCredential(regName, auth.Credential{
			Username: opts.Username,
			Password: opts.Password,
		}),
	}

	return authClient, diags
}

func parseRepoRemoveArgs(args []string) (string, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	if len(args) != 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Required arguments are :[name]",
			Detail:   fmt.Sprintf("Got %s", args),
		})
		return "", diags
	}
	return args[0], diags

}

func RemoveRepo(envSettings *settings.EnvSettings, viewDef *view.ViewArgs, args []string) {
	name, diags := parseRepoRemoveArgs(args)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return
	}

	repos, diags := configs.DecodeRepos(envSettings.RepositoryConfig)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return
	}

	if _, ok := repos[name]; !ok {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Repository doesn't exist",
			Detail:   fmt.Sprintf("Repository %s doesn't exist", name),
		})
		v.DiagPrinter(diags, viewDef)
		return
	}

	delete(repos, name)
	f := hclwrite.NewEmptyFile()
	body := f.Body()
	for _, value := range repos {
		body.AppendBlock(generateBlockFromValue(value))
	}

	repoCacheFile, err := os.OpenFile(envSettings.RepositoryConfig, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't create file",
			Detail:   fmt.Sprint(err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return
	}

	_, err = f.WriteTo(repoCacheFile)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't write to file",
			Detail:   fmt.Sprint(err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return
	}

}

func ListRepos(envSettings *settings.EnvSettings, viewDef *view.ViewArgs, args []string) {

	repos, diags := configs.DecodeRepos(envSettings.RepositoryConfig)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return
	}
	fmt.Println("Name \t\tURL")
	for _, value := range repos {
		fmt.Printf("%s \t\t%s\n", value.Name, value.Url)
	}

}
