package storage

import (
	"github.com/hashicorp/hcl/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

)



type Storage struct {
	resourceList map[string][]byte
} 

func New() *Storage{
	return &Storage{make(map[string][]byte)}
}

func (s *Storage) GenSecret(key string,lbs labels) (*v1.Secret, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	if lbs == nil {
		lbs.init()
	}
	lbs.set("owner","kubehcl")

	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   key,
			Labels: lbs.toMap(),
		},
		Type: "kubehcl.sh/module.v1",
		Data: s.resourceList,
	}, diags
}

func (s *Storage) Add(name string, data []byte){
	s.resourceList[name] = data
}


