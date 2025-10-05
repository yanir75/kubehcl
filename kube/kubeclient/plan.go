package kubeclient

import (
	"bytes"
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"helm.sh/helm/v4/pkg/kube"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/openapi3"
	"k8s.io/kubectl/pkg/cmd/diff"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"kubehcl.sh/kubehcl/internal/decode"
	"kubehcl.sh/kubehcl/internal/view"
)

// This function gets all state resources and builds them into kubernetes objects
func (cfg *Config) buildStateResources() (map[string]kube.ResourceList, hcl.Diagnostics) {
	savedMap, diags := cfg.Storage.GetAllStateResources()
	stateMap := make(map[string]kube.ResourceList)

	if diags.HasErrors() {
		return stateMap, diags
	}

	for key, value := range savedMap {
		// if cfg.Storage.Get(key) != nil {
		reader := bytes.NewReader(value)
		savedResource, builderErr := cfg.Client.Build(reader, true)
		if builderErr != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Couldn't build and validate resources",
				Detail:   fmt.Sprintf("Kind: %s", value),
			})
		}
		stateMap[key] = savedResource
		// }
	}
	return stateMap, diags
}

// This function get all the current state of the resources
func (cfg *Config) GetStateResourcesCurrentState() (map[string]kube.ResourceList, hcl.Diagnostics) {
	sMap, diags := cfg.buildStateResources()
	currentStateMap := make(map[string]kube.ResourceList)

	if diags.HasErrors() {
		return currentStateMap, diags
	}
	for key, value := range sMap {
		if len(value) > 1 || len(value) < 1 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Found too many resources",
				Detail:   fmt.Sprintf("resource %s exists more than once or doesn't exist at all in state", key),
			})
		}
		res, buildDiags := cfg.Storage.BuildResourceFromState(value, key, true)

		diags = append(diags, buildDiags...)
		if len(res) > 0 {
			currentStateMap[key] = res
		}
	}

	return currentStateMap, diags
}

func (cfg *Config) BuildResource(resource *decode.DecodedResource) (map[string]kube.ResourceList, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	resourceMap := make(map[string]kube.ResourceList)

	for key, value := range resource.Config {
		wanted, buildDiags := cfg.buildResource(key, value, &resource.DeclRange)
		if buildDiags.HasErrors() {
			diags = append(diags, buildDiags...)
			return resourceMap, diags
		}
		resourceMap[key] = wanted
	}
	return resourceMap, diags
}

func (cfg *Config) buildObject(to, from kube.ResourceList) (*view.CompareResources, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	f := cmdutil.NewFactory(cfg.Settings.RESTClientGetter())
	o, err := f.OpenAPIV3Client()
	var oRoot openapi3.Root

	if err == nil {
		oRoot = openapi3.NewRoot(o)
	} else {
		diags = append(diags,
			&hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  ("warning: OpenAPI V3 Patch is enabled but is unable to be loaded. Will fall back to OpenAPI V2"),
			})
	}

	ioStreams := genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	obj := diff.InfoObject{
		LocalObj:        to[0].Object,
		Info:            from[0],
		Encoder:         scheme.DefaultJSONEncoder(),
		OpenAPIGetter:   f,
		OpenAPIV3Root:   oRoot,
		Force:           false,
		ServerSideApply: false,
		FieldManager:    "kubectl-client-side-apply",
		ForceConflicts:  false,
		IOStreams:       ioStreams,
	}
	cmp := &view.CompareResources{
		Current: obj.Live(),
	}
	if wanted, err := obj.Merged(); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't get current resource state",
			Detail:   fmt.Sprintf("%s", err),
		})
	} else {
		cmp.Wanted = wanted
	}

	return cmp, diags
}

func (cfg *Config) CompareResources(to, from map[string]kube.ResourceList) (map[string]*view.CompareResources, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	cmpMap := make(map[string]*view.CompareResources)
	for key, toVal := range to {
		if fromVal, exists := from[key]; exists {
			if len(toVal) != 1 {
				panic("Shouldn't get here")
			}
			if len(fromVal) < 1 {
				cmpMap[key] = &view.CompareResources{
					Wanted: toVal[0].Object,
				}
			} else {
				cmp, buildDiags := cfg.buildObject(toVal, fromVal)
				diags = append(diags, buildDiags...)
				cmpMap[key] = cmp
			}
		} else {
			cmpMap[key] = &view.CompareResources{
				Wanted: toVal[0].Object,
			}
		}
	}

	for key, fromVal := range from {
		if _, exists := to[key]; !exists {
			cmpMap[key] = &view.CompareResources{
				Current: fromVal[0].Object,
			}
		}
	}
	return cmpMap, diags
}
