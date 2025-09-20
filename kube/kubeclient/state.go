package kubeclient

import (
	"bytes"
	"context"
	"fmt"

	"encoding/json"

	"github.com/hashicorp/hcl/v2"
	"helm.sh/helm/v4/pkg/kube"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/resource"
)

// Get the current state of applied resources
// State is saved as a secret inside kubernetes in the given namespace
// The secret type is kubehcl.sh/module.v1
func (cfg *Config) getState() (map[string][]byte, hcl.Diagnostics) {
	secret, diags := cfg.Storage.GenSecret(cfg.Name, nil)
	client, err := cfg.Client.Factory.KubernetesClientSet()
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't get state secret",
			Detail:   fmt.Sprintf("%s", err),
		})
	}

	if getSecret, getSecretErr := client.CoreV1().Secrets(cfg.Settings.Namespace()).Get(context.Background(), secret.Name, metav1.GetOptions{}); apierrors.IsNotFound(getSecretErr) {
		return nil, diags
	} else if getSecretErr != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Couldn't get secret names %s", secret.Name),
			Detail:   fmt.Sprintf("Unable to retreive secret err: %s", getSecretErr),
		})
		return nil, diags
	} else {
		return getSecret.Data, diags
	}

}

// Delete current state meaning delete the secret that is responsible for the state
// This occurs during uninstall
func (cfg *Config) deleteState() hcl.Diagnostics {
	secret, diags := cfg.Storage.GenSecret(cfg.Name, nil)
	client, err := cfg.Client.Factory.KubernetesClientSet()
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't delete state secret",
			Detail:   fmt.Sprintf("%s", err),
		})
	}

	if deleteSecretErr := client.CoreV1().Secrets(cfg.Settings.Namespace()).Delete(context.Background(), secret.Name, metav1.DeleteOptions{}); apierrors.IsNotFound(deleteSecretErr) {
		return diags
	}

	return diags
}

// Get all resources as bytes from the current state
// All resources are saved as a json format
func (cfg *Config) getAllResourcesFromState() (map[string][]byte, hcl.Diagnostics) {
	data, diags := cfg.getState()
	resourceMap := make(map[string][]byte)
	if len(data) > 0 {
		err := json.Unmarshal(data["release"], &resourceMap)
		if err != nil {
			panic("should not get here: " + err.Error())
		}
	}

	return resourceMap, diags

}

// Update secret willl apply the new storage stored resources and update the secret accordingly
func (cfg *Config) UpdateSecret() hcl.Diagnostics {
	diags := cfg.updatePreviousReleaseData()
	secret, genSecretDiags := cfg.Storage.GenSecret(cfg.Name, nil)
	diags = append(diags, genSecretDiags...)
	client, err := cfg.Client.Factory.KubernetesClientSet()
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't get client",
			Detail:   fmt.Sprintf("%s", err),
		})
	}

	if _, createSecretErr := client.CoreV1().Secrets(cfg.Settings.Namespace()).Create(context.Background(), secret, metav1.CreateOptions{}); apierrors.IsAlreadyExists(createSecretErr) {
		if _, updateSecretErr := client.CoreV1().Secrets(cfg.Settings.Namespace()).Update(context.Background(), secret, metav1.UpdateOptions{}); updateSecretErr != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Couldn't update state secret",
				Detail:   fmt.Sprintf("%s", updateSecretErr),
			})
		}
	} else if createSecretErr != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't create state secret",
			Detail:   fmt.Sprintf("%s", createSecretErr),
		})
	}

	return diags
}

// Get the current state of a specific resource
// This gets all the attributes from the state and adds them to a the resource
// If the resource is not found or got an error an empty list will be returned
// This is to check if the resource matches the configuration in the state or not
func (cfg *Config) getResourceCurrentState(resources kube.ResourceList) (kube.ResourceList, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var resList kube.ResourceList

	if res, err := cfg.Client.Get(resources, false); apierrors.IsNotFound(err) {
		return resList, diags
	} else if err != nil {
		for key := range res {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Couldn't get resource: %s", key),
			})
		}
		return resList, diags
	} else {
		for _, value := range res {
			for _, val := range value {
				var resourceInfo *resource.Info = &resource.Info{}
				refreshErr := resourceInfo.Refresh(val, false)
				resourceInfo.Mapping = &meta.RESTMapping{}
				resourceInfo.Mapping.Resource = val.GetObjectKind().GroupVersionKind().GroupVersion().WithResource("")
				resourceInfo.Mapping.GroupVersionKind = val.GetObjectKind().GroupVersionKind()
				if refreshErr != nil {
					panic("should not get here: " + refreshErr.Error())
				}
				resList = append(resList, resourceInfo)
			}
		}
	}

	return resList, diags
}

// Get current resource from state builds it in order to verify it and apply the resource later
// Getting the resource verifies that the resource doesn't exist or is managed by kubehcl
// Builds the resource from the state this is done to update the current configuration
// This also verifies if the resource exists and was not saved in the kubehcl state in order to not update it
func (cfg *Config) buildResourceFromState(wanted kube.ResourceList, name string) (kube.ResourceList, hcl.Diagnostics) {
	// Get current resource configuration
	// Get the resource configuration from the state
	current, diags := cfg.getResourceCurrentState(wanted)

	saved, savedData := cfg.getAllResourcesFromState()
	diags = append(diags, savedData...)
	if diags.HasErrors() {
		return nil,diags
	}

	reader := bytes.NewReader(saved[name])
	savedResource, builderErr := cfg.Client.Build(reader, true)

	if builderErr != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't build and validate resources",
			Detail:   fmt.Sprintf("Kind: %s", saved[name]),
		})
		return nil, diags
	}

	// We get and check one resource at a time
	if len(current) > 1 || len(savedResource) > 1 || len(wanted) != 1 {
		panic(fmt.Sprintf("Shouldn't get here\ncurrent:%d\nsavedResource:%d\nwanted:%d", len(current), len(savedResource), len(wanted)))
	}

	if len(current) == 1 && len(savedResource) == 0 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Resource already exists but not managed by Kubehcl",
			Detail:   fmt.Sprintf("Kind: %s,\nResource:%s", current[0].Mapping.GroupVersionKind.Kind, current[0].Name),
		})
		cfg.Storage.Delete(current[0].Name)

		return nil, diags
	}
	return savedResource, diags
}

// Updates the previous releases data
func (cfg *Config) updatePreviousReleaseData() hcl.Diagnostics {
	data, diags := cfg.getState()
	previousDataMap := make(map[string]map[string][]byte)
	if len(data) > 0 {
		err := json.Unmarshal(data["previous-releases"], &previousDataMap)
		if err != nil {
			panic("should not get here: " + err.Error())
		}
	}

	resourceMap := make(map[string][]byte)
	if len(data) > 0 {
		err := json.Unmarshal(data["release"], &resourceMap)
		if err != nil {
			panic("should not get here: " + err.Error())
		}
	}

	cfg.Storage.InitPreviousData(previousDataMap)
	cfg.Storage.AddPreviousData(resourceMap)
	return diags
}
