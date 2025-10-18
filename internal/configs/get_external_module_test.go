package configs

import (
	"path/filepath"
	"testing"

	"kubehcl.sh/kubehcl/internal/decode"
)

func Test_PullHttp(t *testing.T) {
	name := "acs-engine-autoscaler"
	fs, diags := pullHttp(&decode.DecodedRepo{
		Name:     "stable",
		Url:      "charts.helm.sh/stable",
		Protocol: "https",
	}, name, "2.2.2", false)

	if diags.HasErrors() {
		t.Errorf("Failed to pull")
	}

	_, err := fs.Open(name + string(filepath.Separator) + "Chart.yaml")

	if err != nil {
		t.Errorf("Failed to open file %s", err)
	}

}
