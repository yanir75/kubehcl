package client

import (
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/spf13/afero"
	"kubehcl.sh/kubehcl/internal/logging"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
)

func serve(t *testing.T){
    	handler := registry.New()
		if err := http.ListenAndServe(":8080", handler); err != nil {
    		t.Errorf("Failed to server %s", err)
    	}

}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	// Open the source file for reading
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close() // Ensure the source file is closed

	// Create the destination file for writing
	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close() // Ensure the destination file is closed

	// Copy the contents from source to destination
	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	// Optionally, copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	err = os.Chmod(dst, sourceInfo.Mode())
	if err != nil {
		return err
	}

	return nil
}
func Test_AddRepoOci(t *testing.T)  {
	logging.SetLogger(false)

	go serve(t)
	opts := &settings.RepoAddOptions{
			Name: "test",
			Url: "localhost:8080/my-repo",
			Protocol: "oci",
			PlainHttp: true,
			
		}
	envs := &settings.EnvSettings{
			RepositoryConfig: "reg.hcl",
			RepositoryCache: "reg.cache",
		}
	
	diags := AddRepo(
		opts,
		envs,
		&view.ViewArgs{},
		[]string{"test","oci://localhost:8080/my-repo"},
	)
	if diags.HasErrors() {
		t.Errorf("Failed to add repository")
	}

	envs = &settings.EnvSettings{
			RepositoryConfig: "reg2.hcl",
			RepositoryCache: "reg.cache",
		}
	
	err := copyFile("reg.hcl","reg2.hcl")
	if err != nil {
		t.Errorf("Failed to copy %s",err)
	}
	diags = Push(envs,&view.ViewArgs{},[]string{"files","test","v1"})

	if diags.HasErrors() {
		t.Errorf("Failed to push")
	}

	appFs,diags := Pull("",envs,&view.ViewArgs{},[]string{"test","v1"},false)
	if diags.HasErrors(){
		t.Errorf("Failed to pull")
	}

	_,err = afero.ReadFile(appFs,"v1/index.hclvars")

	if err != nil {
		t.Errorf("Failed to read index.hclvars %s",err)

	}
	os.Remove("reg.hcl")
	os.Remove("reg2.hcl")


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