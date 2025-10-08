package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"kubehcl.sh/kubehcl/internal/configs"
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
		})
		return "", "", diags
	}
	return args[0], args[1], diags

}

func RepoAdd(opts *settings.RepoAddOptions,defs *settings.EnvSettings,viewDef *view.ViewArgs,args []string){

	name,url,diags := parseRepoAddArgs(args)
	if diags.HasErrors(){
		v.DiagPrinter(diags,viewDef)
		return
	}

	repos,diags := configs.DecodeRepos(defs.RepositoryConfig)
	if diags.HasErrors(){
		v.DiagPrinter(diags,viewDef)
		return
	}

	if val,ok := repos[name]; ok {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: "Repository already exists",
			Detail: fmt.Sprintf("Repository %s already exists",val.Name),
			Subject: &val.DeclRange,
		})
	}

	if diags.HasErrors(){
		v.DiagPrinter(diags,viewDef)
		return
	}
	
	opts.Name = name
	opts.Url = url
	opts.RepoCache = defs.RepositoryCache
	opts.RepoFile = defs.RepositoryConfig
	repo,err := remote.NewRepository(opts.Url)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: "Couldn't load CA certificate",
			Detail: fmt.Sprint(err.Error()),
		})
		v.DiagPrinter(diags,viewDef)
		return
	}


	repo.Client,diags = newClient(opts,repo.Reference.Registry)

	if diags.HasErrors(){
		v.DiagPrinter(diags,viewDef)
		return
	}

	repo.PlainHTTP = opts.PlainHttp
	repo.TagListPageSize = 1
	err = repo.Tags(context.Background(),"",func(tags []string) error {
		// fmt.Printf("%s",tags)
		return nil
	})
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: "Couldn't authenticate",
			Detail: fmt.Sprint(err.Error()),
		})
		v.DiagPrinter(diags,viewDef)
		return
	}

	f := hclwrite.NewEmptyFile()
	body := f.Body()
	block := body.AppendNewBlock("repo",[]string{opts.Name})
	valueOf := reflect.ValueOf(*opts)
	typeOf := reflect.TypeOf(*opts)

	for i:=0; i< valueOf.NumField() ;i++ {
		
		val := valueOf.Field(i).Interface()
		
		switch tt:=val.(type) {
		case int64:
			block.Body().SetAttributeValue(typeOf.Field(i).Name,cty.NumberIntVal(tt))
		case string:
			block.Body().SetAttributeValue(typeOf.Field(i).Name,cty.StringVal(tt))
		case bool:
			block.Body().SetAttributeValue(typeOf.Field(i).Name,cty.BoolVal(tt))
		case time.Duration:
			block.Body().SetAttributeValue(typeOf.Field(i).Name,cty.NumberIntVal(int64(tt.Seconds())))
		default:
			panic("shouldn't get here")
		}
		
	}
	// err = os.MkdirAll(defs.RepositoryConfig, 0555) // Create with read/write/execute for owner, read/execute for others
	// if err != nil {
	// 	fmt.Printf("Error creating directory: %v\n", err)
	// 	return
	// }

	repoCacheFile, err := os.OpenFile(defs.RepositoryConfig, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: "Couldn't create file",
			Detail: fmt.Sprint(err.Error()),
		})
		v.DiagPrinter(diags,viewDef)
		return
	}

	f.WriteTo(repoCacheFile)
	
}




func newClient(opts *settings.RepoAddOptions,regName string) (*auth.Client,hcl.Diagnostics){
	var diags hcl.Diagnostics
	cfg := &tls.Config{
		InsecureSkipVerify: opts.InsecureSkipTLSverify,
	}
	if opts.KeyFile != "" && opts.CertFile != "" {
    	cert, err := tls.LoadX509KeyPair(opts.CertFile, opts.KeyFile)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary: "Couldn't load certificate",
				Detail: fmt.Sprint(err.Error()),
			})
			return nil,diags
		}
		cfg.Certificates = append(cfg.Certificates, cert)
	}
	if opts.CaFile != "" {
		    caCert, err := os.ReadFile(opts.CaFile)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary: "Couldn't load CA certificate",
					Detail: fmt.Sprint(err.Error()),
				})
				return nil,diags
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
	authClient := &auth.Client{
		Client: httpClient,
		Cache:  auth.NewCache(),
		Credential: auth.StaticCredential(regName, auth.Credential{
			Username: opts.Username,
			Password: opts.Password,
		}),
	}

	return authClient,diags
}