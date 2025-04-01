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
	"sync"

	"github.com/hashicorp/hcl/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var mutex sync.Mutex

var SecretType = "kubehcl.sh/module.v1"

type Storage struct {
	resourceList map[string][]byte
}

func New() *Storage {
	return &Storage{make(map[string][]byte)}
}

func (s *Storage) GenSecret(key string, lbs labels) (*v1.Secret, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	if lbs == nil {
		lbs.init()
	}
	lbs.set("owner", "kubehcl")
	releaseMap := make(map[string][]byte)
	data, err := json.Marshal(s.resourceList)
	if err != nil {
		panic("Should not get here: " + err.Error())
	}
	releaseMap["release"] = data
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "kubehcl." + key,
			Labels: lbs.toMap(),
		},
		Type: v1.SecretType(SecretType),
		Data: releaseMap,
	}, diags
}

func (s *Storage) Add(name string, data []byte) {
	mutex.Lock()
	defer mutex.Unlock()
	s.resourceList[name] = data
}

func (s *Storage) Delete(name string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(s.resourceList, name)
}

func (s *Storage) Get(name string) []byte {
	if data, exists := s.resourceList[name]; exists {
		return data
	}
	return nil
}
