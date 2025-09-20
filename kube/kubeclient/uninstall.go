package kubeclient

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"helm.sh/helm/v4/pkg/kube"
	apierrors "k8s.io/apimachinery/pkg/api/errors"


)

// Delete all resources from a given state
func (cfg *Config) DeleteAllResources() (*kube.Result, hcl.Diagnostics) {

	// var wanted kube.ResourceList = kube.ResourceList{}
	// get saved secret which contains the state
	saved, diags := cfg.getAllResourcesFromState()

	var toDelete kube.ResourceList
	for _, value := range saved {
		// if the key doesn't exist in the storage meaning it doesn't exist in configuration so delete it
		// if cfg.Storage.Get(key) == nil {
		reader := bytes.NewReader(value)
		savedResource, builderErr := cfg.Client.Build(reader, true)
		if builderErr != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Couldn't build and validate resources",
				Detail:   fmt.Sprintf("Kind: %s", value),
			})
			return nil, diags
		}
		toDelete = append(toDelete, savedResource...)
		// }
	}
	// delete all managed resources that don't appear in configuration
	if len(toDelete) < 1 {
		return nil, diags
	}

	res, errs := cfg.Client.Delete(toDelete)
	for i, err := range errs {
		if err != nil && !apierrors.IsNotFound(err) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Couldn't delete resource",
				Detail:   fmt.Sprintf("Kind: %s,\nResource:%s\nerr: %s", res.Deleted[i].Mapping.GroupVersionKind.Kind, res.Deleted[i].Name, err.Error()),
			})
		}
	}

	if err := cfg.Client.WaitForDelete(toDelete, cfg.Timeout); err != nil && !apierrors.IsNotFound(err) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't delete resource within the timeout",
			Detail:   fmt.Sprintf("Kind: %s,\nResource:%s\nerr: %s", toDelete[0].Mapping.GroupVersionKind.Kind, toDelete[0].Name, err.Error()),
		})
	}

	diags = append(diags, cfg.deleteState()...)
	return res, diags
}