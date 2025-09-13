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
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hashicorp/hcl/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var mutex sync.Mutex

var SecretType = "kubehcl.sh/module.v1"

// type rMap map[string][]byte
type Storage struct {
	resourceList map[string][]byte
	previousData map[string]map[string][]byte
}

func New() *Storage {
	return &Storage{make(map[string][]byte),make(map[string]map[string][]byte)}
}


func (s *Storage) marshalData()([]byte,[]byte){
	data, err := json.Marshal(s.resourceList)
	if err != nil {
		panic("Should not get here: " + err.Error())
	}
	prevData,err := json.Marshal(s.previousData)
	if err != nil {
		panic("Should not get here: " + err.Error())
	}
	
	return data,prevData
}
// Generate secret from the current resource list in the storage
func (s *Storage) GenSecret(key string, lbs labels) (*v1.Secret, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	if lbs == nil {
		lbs.init()
	}
	lbs.set("owner", "kubehcl")
	releaseMap := make(map[string][]byte)
	data,prevData :=s.marshalData()
	releaseMap["release"] = data
	releaseMap["previous releases"] = prevData
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "kubehcl." + key,
			Labels: lbs.toMap(),
		},
		Type: v1.SecretType(SecretType),
		Data: releaseMap,
	}, diags
}

func (s *Storage) AddPreviousData(data map[string][]byte){
	str := fmt.Sprintf("release-%d",len(s.previousData))
	s.previousData[str] = data
}

func(s *Storage) InitPreviousData(data map[string]map[string][]byte){
	s.previousData = data
}

// Adda resource to the storage
func (s *Storage) Add(name string, data []byte) {
	mutex.Lock()
	defer mutex.Unlock()
	s.resourceList[name] = data
}

// Delete a resource from the storage
func (s *Storage) Delete(name string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(s.resourceList, name)
}

// Get a resource from the storage
func (s *Storage) Get(name string) []byte {
	if data, exists := s.resourceList[name]; exists {
		return data
	}
	return nil
}
