package openapiclient_test

import (
	"testing"

	"kubehcl.sh/kubehcl/pkg/openapiclient"
)

func TestGitHubBuiltins(t *testing.T) {
	c := openapiclient.NewGitHubBuiltins("1.27")
	_, err := c.Paths()
	if err != nil {
		t.Fatal(err)
	}
}
