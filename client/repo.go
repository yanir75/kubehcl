package client

import (
	"context"
	"fmt"
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

	httpClient, diags := configs.NewHttpClient(OptsToRepo(opts))
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return
	}
	_, diags = configs.DoRequest(OptsToRepo(opts), INDEXFILE, httpClient, "")
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

	repo.Client, diags = configs.NewAuthClient(OptsToRepo(opts), repo.Reference.Registry)

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
			Summary:  "Couldn't fetch tags",
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
