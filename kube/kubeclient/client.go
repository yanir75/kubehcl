/*
This file was inspired from https://github.com/helm/helm
This file has been modified from the original version
Changes made to fit kubehcl purposes
This file retains its' original license
// SPDX-License-Identifier: Apache-2.0
Licesne: https://www.apache.org/licenses/LICENSE-2.0
*/
package kubeclient

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"helm.sh/helm/v4/pkg/kube"

	"encoding/json"

	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/resource"
	"kubehcl.sh/kubehcl/internal/decode"
	"kubehcl.sh/kubehcl/kube/kubeclient/storage"
	"kubehcl.sh/kubehcl/kube/syntaxvalidator"
	"kubehcl.sh/kubehcl/settings"
)

type Config struct {
	Settings *settings.EnvSettings
	Client   *kube.Client
	Storage  *storage.Storage
	Name     string
	Timeout  time.Duration
	Version  string
}

// Applies the settings and creates a config to create,destroy and  validate all configuration files
func New(name string, conf *settings.EnvSettings) (*Config, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	cfg := &Config{}
	cfg.Settings = conf
	cfg.Client = kube.New(cfg.Settings.RESTClientGetter())
	cfg.Storage = storage.New()
	cfg.Name = name
	if conf.Timeout < 0 {
		cfg.Timeout = 100 * time.Second
	} else {
		cfg.Timeout = time.Duration(conf.Timeout) * time.Second
	}
	diags = append(diags, cfg.IsReachable()...)
	if !diags.HasErrors() {
		client, err := cfg.Client.Factory.KubernetesClientSet()
		if err != nil {
			panic("Couldn't get client")
		}
		version, err := client.ServerVersion()
		if err != nil {
			panic("Couldn't get version")
		}
		cfg.Version = version.Major + "." + version.Minor
	}
	return cfg, diags
}

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
	} else {
		return getSecret.Data, diags
	}
}

// Delete current state meaning delete the secret that is responsible for the state
// This occurs during destroy
func (cfg *Config) deleteState() hcl.Diagnostics {
	secret, diags := cfg.Storage.GenSecret(cfg.Name, nil)
	client, err := cfg.Client.Factory.KubernetesClientSet()
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't create or update state secret",
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
	current, diags := cfg.getResourceCurrentState(wanted)
	// Get the resource configuration from the state
	saved, savedData := cfg.getAllResourcesFromState()
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
	diags = append(diags, savedData...)

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
	return current, diags
}

// Compare states get the resource from the state and applies the changes
// If the resource does not exist it will simply be created
func (cfg *Config) compareStates(wanted kube.ResourceList, name string) (*kube.Result, hcl.Diagnostics) {

	current, diags := cfg.buildResourceFromState(wanted, name)
	if diags.HasErrors() {
		return &kube.Result{}, diags
	}
	res, err := cfg.Client.Update(current, wanted, false)

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

// Delete resources will delete all resources in the state
func (cfg *Config) DeleteResources() (map[string]bool, *kube.Result, hcl.Diagnostics) {
	saved, diags := cfg.getAllResourcesFromState()
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
		return deleteMap,nil, diags
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

// Update secret willl apply the new storage stored resources and update the secret accordingly
func (cfg *Config) UpdateSecret() hcl.Diagnostics {
	secret, diags := cfg.Storage.GenSecret(cfg.Name, nil)
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

// Build resource build the resource from cty.value type into a json
func (cfg *Config) buildResource(key string, value cty.Value, rg *hcl.Range) (kube.ResourceList, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	data, err := ctyjson.Marshal(value, value.Type())

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't convert resource config to json",
			Detail:   fmt.Sprintf("%s", err),
			Subject:  rg,
		})
	}

	cfg.Storage.Add(key, data)
	reader := bytes.NewReader(data)
	kubeResourceList, buildErr := cfg.Client.Build(reader, true)
	if buildErr != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't build resource",
			Detail:   fmt.Sprintf("%s", buildErr),
			Subject:  rg,
		})
	}
	return kubeResourceList, diags
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

// Validates the configuration yaml to verify it fits kubernetes
func (cfg *Config) Validate(resource *decode.DecodedResource) hcl.Diagnostics {
	var diags hcl.Diagnostics
	for key, value := range resource.Config {
		data, err := ctyjson.Marshal(value, value.Type())
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Couldn't convert resource config to json",
				Detail:   fmt.Sprintf("%s", err),
				Subject:  &resource.DeclRange,
			})
		}
		factory, validatorDiags := syntaxvalidator.New(cfg.Version)
		diags = append(diags, validatorDiags...)
		err = syntaxvalidator.ValidateDocument(data, factory)

		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Resource Failed Validation",
				Detail:   fmt.Sprintf("Resource: %s failed validation\nErrors will be listed below\n%s", key, formatErr(err)),
				Subject:  &resource.DeclRange,
			})
		}
	}
	return diags
}

// Format for diagnostic error
func formatErr(err error) string {
	errStr := strings.Join(strings.Split(err.Error(), ", "), "\n")
	errStr = strings.ReplaceAll(errStr, "[", "")
	return strings.ReplaceAll(errStr, "]", "")
}

/*
Checks if the client is reachable
*/
func (cfg *Config) IsReachable() hcl.Diagnostics {
	var diags hcl.Diagnostics
	if err := cfg.Client.IsReachable(); err != nil {
		return append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Client is not reachable",
			Detail:   fmt.Sprintf("Error: %s", err.Error()),
		})
	}
	return diags
}

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

// Lists all releases of a namespace
// Does that through listing the secrets matching all secrets matching the type
func (cfg *Config) List() ([]string, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	client, err := cfg.Client.Factory.KubernetesClientSet()
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't get secrets",
			Detail:   fmt.Sprintf("%s", err),
		})
	}

	if secretList, getSecretErr := client.CoreV1().Secrets(cfg.Settings.Namespace()).List(context.Background(), metav1.ListOptions{FieldSelector: "type=" + storage.SecretType}); apierrors.IsNotFound(getSecretErr) {
		return nil, diags
	} else {
		var secretNames []string
		for _, secret := range secretList.Items {
			secretNames = append(secretNames, secret.Name)
		}
		return secretNames, diags
	}
}

// Not implemented
func (cfg *Config) Plan(resource *decode.DecodedResource) (kube.ResourceList, kube.ResourceList, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var wantedList, currentList kube.ResourceList
	// fmt.Printf("%s:%d\n",resource.Name,len(resource.Config))
	for key, value := range resource.Config {
		wanted, buildDiags := cfg.buildResource(key, value, &resource.DeclRange)
		wantedList = append(wantedList, wanted...)
		diags = append(diags, buildDiags...)
		current, buildDiags := cfg.buildResourceFromState(wanted, key)
		currentList = append(currentList, current...)
		diags = append(diags, buildDiags...)
	}

	// if len(wantedList) == 0 {
	// 	panic("err")
	// }
	return currentList, wantedList, diags
}
