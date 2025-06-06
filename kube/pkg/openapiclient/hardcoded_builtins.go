/*
// SPDX-License-Identifier: Apache-2.0
This file was copied from https://github.com/kubernetes-sigs/kubectl-validate and retains its' original license: https://www.apache.org/licenses/LICENSE-2.0
*/
package openapiclient

import (
	"embed"
	"fmt"
	"path"
	"strings"

	"k8s.io/client-go/openapi"
	"kubehcl.sh/kubehcl/kube/pkg/openapiclient/groupversion"
)

//go:generate go run sigs.k8s.io/kubectl-validate/cmd/download-builtin-schemas builtins

//go:embed builtins
var hardcodedBuiltins embed.FS

var HardcodedBuiltinVersions []string = func() []string {
	versions, err := hardcodedBuiltins.ReadDir("builtins")
	if err != nil {
		panic(err)
	}

	res := make([]string, 0, len(versions))
	for _, v := range versions {
		res = append(res, v.Name())
	}

	return res
}()

// client which provides hardcoded openapi for known k8s versions
type hardcodedResolver struct {
	version string
}

func NewHardcodedBuiltins(version string) openapi.Client {
	return hardcodedResolver{version: version}
}

func (k hardcodedResolver) Paths() (map[string]openapi.GroupVersion, error) {
	if len(k.version) == 0 {
		return nil, nil
	}

	allVersions, err := hardcodedBuiltins.ReadDir("builtins")
	if err != nil {
		return nil, err
	}

	for _, v := range allVersions {
		if v.Name() == k.version {
			res := map[string]openapi.GroupVersion{}

			apiDir := path.Join("builtins", v.Name(), "api")
			apiListing, _ := hardcodedBuiltins.ReadDir(apiDir)
			for _, v := range apiListing {
				// chop extension
				ext := path.Ext(v.Name())
				version := strings.TrimSuffix(v.Name(), ext)
				res[fmt.Sprintf("api/%s", version)] = groupversion.NewForFile(&hardcodedBuiltins, path.Join(apiDir, v.Name()))
			}

			apisDir := path.Join("builtins", v.Name(), "apis")
			apisListing, _ := hardcodedBuiltins.ReadDir(apisDir)
			for _, g := range apisListing {
				gDir := path.Join(apisDir, g.Name())
				vs, err := hardcodedBuiltins.ReadDir(gDir)
				if err != nil {
					return nil, err
				}

				for _, v := range vs {
					// chop extension
					ext := path.Ext(v.Name())
					version := strings.TrimSuffix(v.Name(), ext)
					res[fmt.Sprintf("apis/%s/%s", g.Name(), version)] = groupversion.NewForFile(&hardcodedBuiltins, path.Join(gDir, v.Name()))
				}
			}

			return res, nil
		}
	}

	return nil, fmt.Errorf("couldn't find hardcoded schemas for version %s", k.version)
}
