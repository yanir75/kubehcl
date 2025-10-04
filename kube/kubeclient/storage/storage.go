/*
This file was inspired from https://github.com/helm/helm
This file has been modified from the original version
Changes made to fit kubehcl purposes
This file retains its' original license
// SPDX-License-Identifier: Apache-2.0
Licesne: https://www.apache.org/licenses/LICENSE-2.0
*/
package storage

import (
	"github.com/hashicorp/hcl/v2"
	"helm.sh/helm/v4/pkg/kube"
)

type ResourceMap map[string][]byte
type Storage interface {
	Add(name string, data []byte)
	Delete(name string)
	Get(name string) []byte
	GetAllStateResources() (ResourceMap, hcl.Diagnostics)
	GetResourceCurrentState(resources kube.ResourceList) (kube.ResourceList, hcl.Diagnostics)
	BuildResourceFromState(wanted kube.ResourceList, name string,currentOnly bool) (kube.ResourceList, hcl.Diagnostics)
	DeleteState() hcl.Diagnostics
	UpdateState() hcl.Diagnostics
}
