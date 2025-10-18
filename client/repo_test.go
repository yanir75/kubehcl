package client

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/spf13/afero"
	"kubehcl.sh/kubehcl/internal/logging"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
)

var PORT = 0
var REPOURL = "localhost:{port}/my-repo"

const (
	CONFIGFILENAME     = "reg.hcl.config.filename"
	COPYCONFIGFILENAME = "reg.hcl.config.filename2"
	REPONAME           = "repo"
	TAG                = "v1"
	TESTFOLDER         = "files"
	CHARTURL           = "charts.helm.sh/stable"
	VARSFILE           = "index.hclvars"
)

func serve(t *testing.T, listener net.Listener) {
	handler := registry.New()
	if err := http.Serve(listener, handler); err != nil {
		t.Errorf("Failed to server %s", err)
	}
}

func createListener(t *testing.T) net.Listener {

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Couldn't find a fitting port")
	}
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("Couldn't convert addr to *netTCPAddr")
	}
	PORT = addr.Port
	REPOURL = strings.ReplaceAll(REPOURL, "{port}", fmt.Sprint(PORT))
	return listener
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	// Open the source file for reading
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(){_ = source.Close()}() // Ensure the source file is closed

	// Create the destination file for writing
	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(){_ = destination.Close()}() // Ensure the destination file is closed

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

func Test_Repos(t *testing.T) {
	logging.SetLogger(false)
	l := createListener(t)
	// defer func(){_ = l.Close()}()
	go serve(t, l)

	AddRepoOciT(t)
	PushRepoOciT(t)
	PullRepoOciT(t)
	_ = os.Remove(CONFIGFILENAME)
	_ = os.Remove(COPYCONFIGFILENAME)

}

func AddRepoOciT(t *testing.T) {

	opts := &settings.RepoAddOptions{
		Name:      TAG,
		Url:       REPOURL,
		Protocol:  "oci",
		PlainHttp: true,
	}

	envs := &settings.EnvSettings{
		RepositoryConfig: CONFIGFILENAME,
		RepositoryCache:  CONFIGFILENAME + ".cache",
	}

	diags := AddRepo(
		opts,
		envs,
		&view.ViewArgs{},
		[]string{REPONAME, "oci://" + REPOURL},
	)
	if diags.HasErrors() {
		t.Errorf("Failed to add repository")
	}

}

func PushRepoOciT(t *testing.T) {
	envs := &settings.EnvSettings{
		RepositoryConfig: COPYCONFIGFILENAME,
		RepositoryCache:  "reg.cache",
	}

	err := copyFile(CONFIGFILENAME, COPYCONFIGFILENAME)
	if err != nil {
		t.Errorf("Failed to copy %s", err)
	}
	diags := Push(envs, &view.ViewArgs{}, []string{TESTFOLDER, REPONAME, TAG})

	if diags.HasErrors() {
		t.Errorf("Failed to push")
	}
}

func PullRepoOciT(t *testing.T) {
	envs := &settings.EnvSettings{
		RepositoryConfig: COPYCONFIGFILENAME,
		RepositoryCache:  "reg.cache",
	}
	appFs, diags := Pull("", envs, &view.ViewArgs{}, []string{REPONAME, TAG}, false)
	if diags.HasErrors() {
		t.Errorf("Failed to pull")
	}

	_, err := afero.ReadFile(appFs, TAG+afero.FilePathSeparator+VARSFILE)

	if err != nil {
		t.Errorf("Failed to read index.hclvars %s", err)

	}
}

func Test_AddRepoHttp(t *testing.T) {
	diags := AddRepo(
		&settings.RepoAddOptions{
			Name:     REPONAME,
			Url:      CHARTURL,
			Protocol: "https",
		},
		&settings.EnvSettings{
			RepositoryConfig: CONFIGFILENAME,
			RepositoryCache:  "reg.cache",
		},
		&view.ViewArgs{},
		[]string{REPONAME, "https://" + CHARTURL},
	)
	if diags.HasErrors() {
		t.Errorf("Failed to add repository")
	}
	_ = os.Remove(CONFIGFILENAME)
}
