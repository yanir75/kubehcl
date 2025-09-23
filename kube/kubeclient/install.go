package kubeclient

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"helm.sh/helm/v4/pkg/kube"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubehcl.sh/kubehcl/internal/decode"
)

func (cfg *Config) VerifyInstall(createNamespace bool) hcl.Diagnostics {
	client, err := cfg.Client.Factory.KubernetesClientSet()
	if err != nil {
		panic("Couldn't get client")
	}
	var diags hcl.Diagnostics

	if createNamespace {
		ns := &v1.Namespace{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Namespace",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: cfg.Settings.Namespace(),
				Labels: map[string]string{
					"name": cfg.Settings.Namespace(),
				},
			},
		}

		if _, err := client.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{}); err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Couldn't create namepsace %s", cfg.Settings.Namespace()),
				Detail:   err.Error(),
			})
		}
	} else {
		return cfg.validateNamespace()
	}
	return diags
}

// Delete resources will delete all resources in the state that are not in the configuration files
func (cfg *Config) DeleteResources() (map[string]bool, *kube.Result, hcl.Diagnostics) {
	saved, diags := cfg.Storage.GetAllStateResources()
	var toDelete kube.ResourceList
	deleteMap := make(map[string]bool)
	for key, value := range saved {
		if cfg.Storage.Get(key) == nil {
			reader := bytes.NewReader(value)
			savedResource, builderErr := cfg.Client.Build(reader, true)
			if builderErr != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Couldn't build and validate resources",
					Detail:   fmt.Sprintf("Kind: %s", value),
				})
				return deleteMap, nil, diags
			}
			toDelete = append(toDelete, savedResource...)
			deleteMap[key] = true
		}
	}

	if len(toDelete) < 1 {
		return deleteMap, nil, diags
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
	return deleteMap, res, diags
}

// Compare states get the resource from the state and applies the changes
// If the resource does not exist it will simply be created
func (cfg *Config) compareStates(wanted kube.ResourceList, name string) (*kube.Result, hcl.Diagnostics) {
	// if cfg.StorageKind == "stateless"
	current, diags := cfg.Storage.BuildResourceFromState(wanted, name)
	if diags.HasErrors() {
		return &kube.Result{}, diags
	}
	res, err := cfg.Client.Update(current, wanted)

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't update resource",
			Detail:   fmt.Sprintf("Kind: %s,\nResource:%s\nerr: %s", wanted[0].Mapping.GroupVersionKind.Kind, wanted[0].Name, err.Error()),
		})
		return res, diags
	}

	if err := cfg.Client.Wait(wanted, cfg.Timeout); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Resource is not ready within the timeout",
			Detail:   fmt.Sprintf("Kind: %s,\nResource:%s\nerr: %s", wanted[0].Mapping.GroupVersionKind.Kind, wanted[0].Name, err.Error()),
		})
	}

	return res, diags
}

// Create updates the current state to fit the new configuration and updates the current state accordingly
func (cfg *Config) Create(resource *decode.DecodedResource) (*kube.Result, hcl.Diagnostics) {

	var diags hcl.Diagnostics
	var results *kube.Result = &kube.Result{}
	for key, value := range resource.Config {

		kubeResourceList, buildDiags := cfg.buildResource(key, value, &resource.DeclRange)
		diags = append(diags, buildDiags...)
		res, updateDiags := cfg.compareStates(kubeResourceList, key)
		if !updateDiags.HasErrors() {
			results.Created = append(results.Created, res.Created...)
			results.Updated = append(results.Updated, res.Updated...)
			results.Deleted = append(results.Deleted, res.Deleted...)
		}
		diags = append(diags, updateDiags...)
	}

	for _, diag := range diags {
		diag.Subject = &resource.DeclRange
	}

	return results, diags

}
