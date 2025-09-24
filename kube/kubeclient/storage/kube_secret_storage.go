package storage

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"encoding/json"

	"github.com/hashicorp/hcl/v2"
	"helm.sh/helm/v4/pkg/kube"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/resource"
)

var mutex sync.Mutex

var SecretType = "kubehcl.sh/module.v1"

type KubeSecretStorage struct {
	resourceMap  ResourceMap
	previousData map[string]ResourceMap
	client       *kube.Client
	name         string
	namespace    string
	storageKind  string
	stateData map[string][]byte
}

func New(client *kube.Client, name string, namespace string, storageKind string) (Storage, hcl.Diagnostics) {
	kubeStorage := &KubeSecretStorage{
		resourceMap:  make(map[string][]byte),
		previousData: make(map[string]ResourceMap),
		client:       client,
		name:         name,
		namespace:    namespace,
		storageKind:  storageKind,
	}
	prevStorageKind, diags := kubeStorage.getStorageKind()
	if diags.HasErrors() {
		return nil, diags
	}

	if prevStorageKind != "" && storageKind != prevStorageKind {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  fmt.Sprintf("Storage kind has changed from %s to %s", prevStorageKind, storageKind),
		})
	}
	return kubeStorage, diags
}

func (s *KubeSecretStorage) marshalData() ([]byte, []byte) {
	data, err := json.Marshal(s.resourceMap)
	if err != nil {
		panic("Should not get here: " + err.Error())
	}
	prevData, err := json.Marshal(s.previousData)
	if err != nil {
		panic("Should not get here: " + err.Error())
	}

	return data, prevData
}

// Generate secret from the current resource list in the storage
func (s *KubeSecretStorage) genSecret(key string, lbs labels) (*v1.Secret, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	if lbs == nil {
		lbs.init()
	}
	lbs.set("owner", "kubehcl")
	releaseMap := make(map[string][]byte)
	data, prevData := s.marshalData()
	releaseMap["release"] = data
	releaseMap["previous-releases"] = prevData
	releaseMap["storage-kind"] = []byte(s.storageKind)

	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "kubehcl." + key,
			Labels: lbs.toMap(),
		},
		Type: v1.SecretType(SecretType),
		Data: releaseMap,
	}, diags
}

func (s *KubeSecretStorage) AddPreviousData(data map[string][]byte) {
	str := fmt.Sprintf("release-%d", len(s.previousData))
	s.previousData[str] = data
}

func (s *KubeSecretStorage) InitPreviousData(data map[string]ResourceMap) {
	s.previousData = data
}

// // Adda resource to the storage
func (s *KubeSecretStorage) Add(name string, data []byte) {
	mutex.Lock()
	defer mutex.Unlock()
	s.resourceMap[name] = data
}

// // Delete a resource from the storage
func (s *KubeSecretStorage) Delete(name string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(s.resourceMap, name)
}

// Get a resource from the storage
func (s *KubeSecretStorage) Get(name string) []byte {
	if data, exists := s.resourceMap[name]; exists {
		return data
	}
	return nil
}

// Get the current state of applied resources
// State is saved as a secret inside kubernetes in the given namespace
// The secret type is kubehcl.sh/module.v1
func (s *KubeSecretStorage) getState() (map[string][]byte, hcl.Diagnostics) {
	if s.stateData != nil {

		return s.stateData,hcl.Diagnostics{}
	}
	secret, diags := s.genSecret(s.name, nil)
	client, err := s.client.Factory.KubernetesClientSet()
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't get state secret",
			Detail:   fmt.Sprintf("%s", err),
		})
	}

	if getSecret, getSecretErr := client.CoreV1().Secrets(s.namespace).Get(context.Background(), secret.Name, metav1.GetOptions{}); apierrors.IsNotFound(getSecretErr) {
		return nil, diags
	} else if getSecretErr != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Couldn't get secret names %s", secret.Name),
			Detail:   fmt.Sprintf("Unable to retreive secret err: %s", getSecretErr),
		})
		return nil, diags
	} else {
		s.stateData = getSecret.Data
		return getSecret.Data, diags
	}

}

// Delete current state meaning delete the secret that is responsible for the state
// This occurs during uninstall
func (s *KubeSecretStorage) DeleteState() hcl.Diagnostics {
	secret, diags := s.genSecret(s.name, nil)
	client, err := s.client.Factory.KubernetesClientSet()
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't delete state secret",
			Detail:   fmt.Sprintf("%s", err),
		})
	}

	if deleteSecretErr := client.CoreV1().Secrets(s.namespace).Delete(context.Background(), secret.Name, metav1.DeleteOptions{}); apierrors.IsNotFound(deleteSecretErr) {
		return diags
	}

	return diags
}

// Get all resources as bytes from the current state
// All resources are saved as a json format
func (s *KubeSecretStorage) GetAllStateResources() (ResourceMap, hcl.Diagnostics) {
	if s.storageKind == "stateless" {
		return make(map[string][]byte), hcl.Diagnostics{}
	}
	data, diags := s.getState()
	resourceMap := make(map[string][]byte)
	if len(data) > 0 {
		err := json.Unmarshal(data["release"], &resourceMap)
		if err != nil {
			panic("should not get here: " + err.Error())
		}
	}

	return resourceMap, diags

}

func (s *KubeSecretStorage) getStorageKind() (string, hcl.Diagnostics) {
	data, diags := s.getState()
	storageKind := ""
	if len(data) > 0 {
		storageKind = string(data["storage-kind"])
	}

	return storageKind, diags

}

// Updates the previous releases data
func (s *KubeSecretStorage) updatePreviousReleaseData() hcl.Diagnostics {
	data, diags := s.getState()
	var previousDataMap = make(map[string]ResourceMap)

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

	s.InitPreviousData(previousDataMap)
	s.AddPreviousData(resourceMap)
	return diags
}

// Update secret willl apply the new storage stored resources and update the secret accordingly
func (s *KubeSecretStorage) UpdateState() hcl.Diagnostics {

	diags := s.updatePreviousReleaseData()
	secret, genSecretDiags := s.genSecret(s.name, nil)
	diags = append(diags, genSecretDiags...)
	client, err := s.client.Factory.KubernetesClientSet()
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't get client",
			Detail:   fmt.Sprintf("%s", err),
		})
	}

	if _, createSecretErr := client.CoreV1().Secrets(s.namespace).Create(context.Background(), secret, metav1.CreateOptions{}); apierrors.IsAlreadyExists(createSecretErr) {
		if _, updateSecretErr := client.CoreV1().Secrets(s.namespace).Update(context.Background(), secret, metav1.UpdateOptions{}); updateSecretErr != nil {
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
func (s *KubeSecretStorage) GetResourceCurrentState(resources kube.ResourceList) (kube.ResourceList, hcl.Diagnostics) {
	if s.storageKind == "stateless" {
		return kube.ResourceList{}, hcl.Diagnostics{}
	}
	var diags hcl.Diagnostics
	var resList kube.ResourceList

	if res, err := s.client.Get(resources, false); apierrors.IsNotFound(err) {
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
				var resourceInfo = &resource.Info{}
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
func (s *KubeSecretStorage) BuildResourceFromState(wanted kube.ResourceList, name string) (kube.ResourceList, hcl.Diagnostics) {
	// Get current resource configuration
	// Get the resource configuration from the state
	if s.storageKind == "stateless" {
		return wanted, hcl.Diagnostics{}
	}
	current, diags := s.GetResourceCurrentState(wanted)

	saved, savedData := s.GetAllStateResources()
	diags = append(diags, savedData...)
	if diags.HasErrors() {
		return nil, diags
	}

	reader := bytes.NewReader(saved[name])
	savedResource, builderErr := s.client.Build(reader, true)

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
		s.Delete(current[0].Name)

		return nil, diags
	}
	return savedResource, diags
}
