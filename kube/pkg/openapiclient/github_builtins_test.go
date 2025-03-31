/* 
// SPDX-License-Identifier: Apache-2.0
This file was copied from https://github.com/kubernetes-sigs/kubectl-validate and retains its' original license: https://www.apache.org/licenses/LICENSE-2.0
*/
package openapiclient_test

import (
	"testing"

	"kubehcl.sh/kubehcl/kube/pkg/openapiclient"
)

func TestGitHubBuiltins(t *testing.T) {
	c := openapiclient.NewGitHubBuiltins("1.27")
	_, err := c.Paths()
	if err != nil {
		t.Fatal(err)
	}
}

