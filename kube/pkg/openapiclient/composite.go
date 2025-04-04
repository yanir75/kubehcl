/*
// SPDX-License-Identifier: Apache-2.0
This file was copied from https://github.com/kubernetes-sigs/kubectl-validate and retains its' original license: https://www.apache.org/licenses/LICENSE-2.0
*/
package openapiclient

import (
	"errors"

	"k8s.io/client-go/openapi"
	"kubehcl.sh/kubehcl/kube/pkg/openapiclient/groupversion"
)

type compositeClient struct {
	clients []openapi.Client
}

// client which tries multiple clients in a priority order for an openapi spec
func NewComposite(clients ...openapi.Client) openapi.Client {
	return compositeClient{clients: clients}
}

func (c compositeClient) Paths() (map[string]openapi.GroupVersion, error) {
	merged := map[string][]openapi.GroupVersion{}
	var allErrors []error
	for _, client := range c.clients {
		paths, err := client.Paths()
		if err != nil {
			allErrors = append(allErrors, err)
			continue
		}
		for k, v := range paths {
			merged[k] = append(merged[k], v)
		}
	}
	composite := map[string]openapi.GroupVersion{}
	for k, v := range merged {
		composite[k] = groupversion.NewForComposite(v...)
	}

	var er error
	if len(allErrors) == 1 {
		er = allErrors[0]
	} else if len(allErrors) > 0 {
		er = errors.Join(allErrors...)
	}

	return composite, er
}
