package client

import (
	"net/http"
	"os"
	"testing"

	"github.com/google/go-containerregistry/pkg/registry"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
)

func serve(t *testing.T){
    	handler := registry.New()
		if err := http.ListenAndServe(":8080", handler); err != nil {
    		t.Errorf("Failed to server %s", err)
    	}

}

func Test_AddRepoOci(t *testing.T)  {

	go serve(t)
	diags := AddRepo(
		&settings.RepoAddOptions{
			Name: "test",
			Url: "localhost:8080/my-repo",
			Protocol: "oci",
			PlainHttp: true,
			
		},
		&settings.EnvSettings{
			RepositoryConfig: "reg.config",
			RepositoryCache: "reg.cache",
		},
		&view.ViewArgs{},
		[]string{"test","oci://localhost:8080/my-repo"},
	)
	if diags.HasErrors() {
		t.Errorf("Failed to add repository")
	}
	os.Remove("reg.config")

}

func Test_AddRepoHttp(t *testing.T) {
	diags := AddRepo(
		&settings.RepoAddOptions{
			Name: "test",
			Url: "charts.helm.sh/stable",
			Protocol: "https",			
		},
		&settings.EnvSettings{
			RepositoryConfig: "reg.config",
			RepositoryCache: "reg.cache",
		},
		&view.ViewArgs{},
		[]string{"test","https://charts.helm.sh/stable"},
	)
	if diags.HasErrors() {
		t.Errorf("Failed to add repository")
	}
	os.Remove("reg.config")
}