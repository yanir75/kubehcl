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
}

func New(name string) (*Config, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	cfg := &Config{}
	cfg.Settings = settings.New()
	cfg.Client = kube.New(cfg.Settings.RESTClientGetter())
	cfg.Storage = storage.New()
	cfg.Name = name
	diags = append(diags, cfg.IsReachable()...)

	return cfg, diags
}

// func (cfg *Config) Create() hcl.Diagnostics{

// }
func (cfg *Config) getState() (map[string][]byte, hcl.Diagnostics) {
	secret, diags := cfg.Storage.GenSecret(cfg.Name, nil)
	client, err := cfg.Client.Factory.KubernetesClientSet()
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't create or update state secret",
			Detail:   fmt.Sprintf("%s", err),
		})
	}

	if getSecret, getSecretErr := client.CoreV1().Secrets(cfg.Settings.Namespace()).Get(context.Background(), secret.Name, metav1.GetOptions{}); apierrors.IsNotFound(getSecretErr) {
		return nil, diags
	} else {
		return getSecret.Data, diags
	}
}

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

func (cfg *Config) buildResourceFromState(wanted kube.ResourceList,name string)(kube.ResourceList,hcl.Diagnostics){
	current, diags := cfg.getResourceCurrentState(wanted)
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
	return current,diags
}

func (cfg *Config) compareStates(wanted kube.ResourceList, name string) (*kube.Result, hcl.Diagnostics) {

	current,diags := cfg.buildResourceFromState(wanted,name)
	if diags.HasErrors() {
		return &kube.Result{},diags
	}
	res, err := cfg.Client.Update(current, wanted, false)

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't update resource",
			Detail:   fmt.Sprintf("Kind: %s,\nResource:%s\nerr: %s", wanted[0].Mapping.GroupVersionKind.Kind, wanted[0].Name, err.Error()),
		})
	}

	cfg.Client.Wait(wanted, 100)
	return res, diags
}

func (cfg *Config) DeleteResources() (*kube.Result, hcl.Diagnostics) {
	var wanted kube.ResourceList = kube.ResourceList{}
	saved, diags := cfg.getAllResourcesFromState()
	var toDelete kube.ResourceList
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
				return nil, diags
			}
			toDelete = append(toDelete, savedResource...)
		}
	}

	res, err := cfg.Client.Update(toDelete, wanted, false)
	if err != nil {
		for _, res := range toDelete {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Couldn't delete resource",
				Detail:   fmt.Sprintf("Kind: %s,\nResource:%s\nerr: %s", res.Mapping.GroupVersionKind.Kind, res.Name, err.Error()),
			})
		}
	}

	cfg.Client.Wait(wanted, 100)
	return res, diags
}

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
func (cfg *Config) buildResource (key string ,value cty.Value,rg *hcl.Range) (kube.ResourceList,hcl.Diagnostics) {
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
	return kubeResourceList,diags
}

func (cfg *Config) Create(resource *decode.DecodedResource) (*kube.Result, hcl.Diagnostics) {

	var diags hcl.Diagnostics
	var results *kube.Result = &kube.Result{}
	for key, value := range resource.Config {


		kubeResourceList,buildDiags := cfg.buildResource(key,value,&resource.DeclRange)
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
		factory, validatorDiags := syntaxvalidator.New()
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

func (cfg *Config) DeleteAllResources() (*kube.Result, hcl.Diagnostics) {
	var wanted kube.ResourceList = kube.ResourceList{}
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
	res, err := cfg.Client.Update(toDelete, wanted, false)
	if err != nil {
		for _, res := range toDelete {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Couldn't delete resource",
				Detail:   fmt.Sprintf("Kind: %s,\nResource:%s\nerr: %s", res.Mapping.GroupVersionKind.Kind, res.Name, err.Error()),
			})
		}
	}

	cfg.Client.Wait(wanted, 100)
	cfg.deleteState()
	return res, diags
}

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


func (cfg *Config) Plan(resource *decode.DecodedResource) (kube.ResourceList,kube.ResourceList, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var wantedList,currentList kube.ResourceList
	// fmt.Printf("%s:%d\n",resource.Name,len(resource.Config))
	for key, value := range resource.Config {
		wanted,buildDiags := cfg.buildResource(key,value,&resource.DeclRange)
		wantedList = append(wantedList, wanted...)
		diags = append(diags, buildDiags...)
		current,buildDiags := cfg.buildResourceFromState(wanted,key)
		currentList = append(currentList, current...)
		diags = append(diags, buildDiags...)
	}
	// if len(wantedList) == 0 {
	// 	panic("err")
	// }
	return currentList,wantedList,diags
}
