/*
This file was inspired from https://github.com/kubernetes-sigs/kubectl-validate
This file has been modified from the original version
Changes made to fit kubehcl purposes
This file retains its' original license
// SPDX-License-Identifier: Apache-2.0
Licesne: https://www.apache.org/licenses/LICENSE-2.0
*/
package syntaxvalidator

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/hashicorp/hcl/v2"
	"k8s.io/apiextensions-apiserver/pkg/apiserver"
	"k8s.io/apiextensions-apiserver/pkg/registry/customresourcedefinition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"kubehcl.sh/kubehcl/kube/pkg/openapiclient"
	"kubehcl.sh/kubehcl/kube/pkg/validator"
)

func New(version string) (*validator.Validator, hcl.Diagnostics) {
	var schemaPatchesFs, localSchemasFs fs.FS

	var localCRDsFileSystems []fs.FS
	var diags hcl.Diagnostics
	// tool fetches openapi in the following priority order:
	// factory, err :=
	factory, err := validator.New(
		openapiclient.NewOverlay(
			// apply user defined patches on top of the final schema
			openapiclient.PatchLoaderFromDirectory(schemaPatchesFs),
			openapiclient.NewComposite(
				// consult local OpenAPI
				openapiclient.NewLocalSchemaFiles(localSchemasFs),
				// consult local CRDs
				openapiclient.NewLocalCRDFiles(localCRDsFileSystems...),
				openapiclient.NewOverlay(
					// Hand-written hardcoded patches.
					openapiclient.HardcodedPatchLoader(version),
					// try cluster for schemas first, if they are not available
					// then fallback to hardcoded or builtin schemas
					openapiclient.NewFallback(
						// contact connected cluster for any schemas. (should this be opt-in?)
						// openapiclient.NewKubeConfig(),
						// try hardcoded builtins first, if they are not available
						// fall back to GitHub builtins
						openapiclient.NewFallback(
							// schemas for known k8s versions are scraped from GH and placed here
							openapiclient.NewHardcodedBuiltins(version),
							// check github for builtins not hardcoded.
							// subject to rate limiting. should use a diskcache
							// since etag requests are not limited
							// openapiclient.NewGitHubBuiltins(version),
						)),
				),
			),
		),
	)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: "Couldn't build validator",
			Detail:  fmt.Sprintf("Kubernetes syntax won't be validated %s",err),
			
		})
		return nil, diags
	}
	return factory, diags

}

func ValidateDocument(document []byte, resolver *validator.Validator) error {
	gvk, parsed, err := resolver.Parse(document)
	if gvk.Group == "apiextensions.k8s.io" && gvk.Kind == "CustomResourceDefinition" {
		// CRD spec contains an infinite loop which is not supported by K8s
		// OpenAPI-based validator. Use the handwritten validation based upon
		// native types for CRD files. There are no other recursive schemas to my
		// knowledge, and any schema defined in CRD cannot be recursive.
		// Long term goal is to remove this once k8s upstream has better
		// support for validating against spec.Schema for native types.
		obj, _, err := serializer.NewCodecFactory(apiserver.Scheme).UniversalDecoder().Decode(document, nil, nil)
		if err != nil {
			return err
		}

		strat := customresourcedefinition.NewStrategy(apiserver.Scheme)
		rest.FillObjectMetaSystemFields(obj.(metav1.Object))
		return rest.BeforeCreate(strat, request.WithNamespace(context.TODO(), ""), obj)
	} else if err != nil {
		return err
	}
	return resolver.Validate(parsed)
}
