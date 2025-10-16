package configs

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/google/go-containerregistry/pkg/registry"
	"kubehcl.sh/kubehcl/internal/decode"
)
func serve(t *testing.T){
    	handler := registry.New()
		if err := http.ListenAndServe(":8080", handler); err != nil {
    		t.Errorf("Failed to server %s", err)
    	}

}
func Test_PullHttp(t *testing.T) {
	name := "acs-engine-autoscaler"
	fs,diags := pullHttp(&decode.DecodedRepo{
		Name: "stable",
		Url: "charts.helm.sh/stable",
		Protocol: "https",
	},name,"2.2.2",false)

	if diags.HasErrors() {
		t.Errorf("Failed to pull")
	}

	_,err := fs.Open(name+string(filepath.Separator)+"Chart.yaml")
	
	if err != nil {
		t.Errorf("Failed to open file %s",err)
	}

}
