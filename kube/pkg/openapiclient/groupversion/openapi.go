/* 
// SPDX-License-Identifier: Apache-2.0
This file was copied from https://github.com/kubernetes-sigs/kubectl-validate and retains its' original license: https://www.apache.org/licenses/LICENSE-2.0
*/
package groupversion

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/openapi"
	"k8s.io/kube-openapi/pkg/spec3"
)

type OpenApiGroupVersion struct {
	*spec3.OpenAPI
}

func (gv *OpenApiGroupVersion) Schema(contentType string) ([]byte, error) {
	if strings.ToLower(contentType) != runtime.ContentTypeJSON {
		return nil, fmt.Errorf("only application/json content type is supported")
	}
	return json.Marshal(gv.OpenAPI)
}

func (gv *OpenApiGroupVersion) ServerRelativeURL() string {
	return ""
}

func NewForOpenAPI(spec *spec3.OpenAPI) openapi.GroupVersion {
	return &OpenApiGroupVersion{spec}
}

