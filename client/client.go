package client

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"helm.sh/helm/v4/pkg/kube"

	ctyjson "github.com/zclconf/go-cty/cty/json"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/resource"
	"kubehcl.sh/kubehcl/cli"
	"kubehcl.sh/kubehcl/client/storage"
	"kubehcl.sh/kubehcl/internal/decode"
)


type Config struct {
	Settings *cli.EnvSettings
	Client *kube.Client
	Storage *storage.Storage 
	Name string
}


func New() *Config {

	cfg := &Config{}
	cfg.Settings = cli.New()
	cfg.Client = kube.New(cfg.Settings.RESTClientGetter())
	cfg.Storage = storage.New()
	cfg.Name = "test"
	return cfg
}


// func (cfg *Config) Create() hcl.Diagnostics{

// }
func (cfg *Config) getState(module string) (map[string][]byte,hcl.Diagnostics){
	secret ,diags :=cfg.Storage.GenSecret(module,nil)
	client, err :=cfg.Client.Factory.KubernetesClientSet()
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: "Couldn't create or update state secret",
			Detail: fmt.Sprintf("%s",err),
		})
	}

	if getSecret, getSecretErr :=client.CoreV1().Secrets(cfg.Settings.Namespace()).Get(context.Background(),secret.Name,metav1.GetOptions{}); apierrors.IsNotFound(getSecretErr) {
		return nil,diags
	} else {
		return getSecret.Data,diags
	}

}

func (cfg *Config) getResourceCurrentState(resources kube.ResourceList) (kube.ResourceList,hcl.Diagnostics){
	var diags hcl.Diagnostics
	var resList kube.ResourceList

	if res,err :=cfg.Client.Get(resources,false); apierrors.IsNotFound(err){
		return resList,diags
	} else if err != nil {
		for key := range res {		
			diags = append(diags,&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary: fmt.Sprintf("Couldn't get resource: %s",key),
			})
		}
		return resList,diags
	} else {
		for _,value := range res {
			for _,val := range value{
				var resourceInfo *resource.Info = &resource.Info{}
				refreshErr :=resourceInfo.Refresh(val,false)
				resourceInfo.Mapping = &meta.RESTMapping{}
				resourceInfo.Mapping.Resource = val.GetObjectKind().GroupVersionKind().GroupVersion().WithResource("")
				resourceInfo.Mapping.GroupVersionKind = val.GetObjectKind().GroupVersionKind()
				if refreshErr != nil {
					panic("should not get here: "+refreshErr.Error())
				} 
				resList = append(resList, resourceInfo)
			}
		}
	}

	return resList,diags
}

func (cfg *Config) compareStates(wanted kube.ResourceList,module string,name string) (*kube.Result,hcl.Diagnostics){
	current,diags :=cfg.getResourceCurrentState(wanted)
	saved,savedData := cfg.getState(module)
	reader := bytes.NewReader(saved[name])
	savedResource,_ :=cfg.Client.Build(reader,true)
	diags = append(diags, savedData...)

	if len(current) > 1 || len(savedResource) > 1 || len(wanted) !=1 {
		panic("Shouldn't get here")
	} 

	if len(current) ==1 && len(savedResource) == 0{
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: "Resource already exists but not managed by Kubehcl",
			Detail: fmt.Sprintf("Kind: %s,\nResource:%s",current[0].Mapping.GroupVersionKind.Kind,current[0].Name,),
		})
		return nil,diags
	}

	res,err := cfg.Client.Update(current,wanted,false)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: "Couldn't update resource",
			Detail: fmt.Sprintf("Kind: %s,\nResource:%s\nerr: %s",wanted[0].Mapping.GroupVersionKind.Kind,wanted[0].Name,err.Error()),
		})
	}

	cfg.Client.Wait(wanted,100)
	return res,diags
}

func (cfg *Config) UpdateSecret(module string) hcl.Diagnostics{
	secret ,diags :=cfg.Storage.GenSecret(module,nil)
	client, err :=cfg.Client.Factory.KubernetesClientSet()
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: "Couldn't get client",
			Detail: fmt.Sprintf("%s",err),
		})
	}

	if _,createSecretErr :=client.CoreV1().Secrets(cfg.Settings.Namespace()).Create(context.Background(),secret,metav1.CreateOptions{}); apierrors.IsAlreadyExists(createSecretErr){
		if _,updateSecretErr :=client.CoreV1().Secrets(cfg.Settings.Namespace()).Update(context.Background(),secret,metav1.UpdateOptions{}); updateSecretErr !=nil{
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary: "Couldn't update state secret",
				Detail: fmt.Sprintf("%s",updateSecretErr),
			})
		} 
	} else if createSecretErr != nil{
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: "Couldn't create state secret",
			Detail: fmt.Sprintf("%s",createSecretErr),
		})
	}

	return diags
}

func (cfg *Config) Create(resource *decode.DecodedResource) (*kube.Result,hcl.Diagnostics){

	var diags hcl.Diagnostics
	var results *kube.Result = &kube.Result{}
	for key,value := range resource.Config {
		storageKey := strings.ReplaceAll(key,"[","(")
		storageKey = strings.ReplaceAll(storageKey,"]",")")
		data,err :=ctyjson.Marshal(value,value.Type())
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary: "Couldn't convert resource config to json",
				Detail: fmt.Sprintf("%s",err),
				Subject: &resource.DeclRange,
			})
		}
		cfg.Storage.Add(storageKey,data)
		reader := bytes.NewReader(data)
		kubeResourceList,buildErr :=cfg.Client.Build(reader,true)
		if buildErr != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary: "Couldn't build resource",
				Detail: fmt.Sprintf("%s",buildErr),
				Subject: &resource.DeclRange,
			})
		}
		res,updateDiags := cfg.compareStates(kubeResourceList,cfg.Name,storageKey)
		if !updateDiags.HasErrors(){
			results.Created = append(results.Created, res.Created...)
			results.Updated = append(results.Updated, res.Updated...)
			results.Deleted = append(results.Deleted, res.Deleted...)
		}
		diags = append(diags,updateDiags...)
	}
	
	// secretDiags :=cfg.UpdateSecret(cfg.Name)
	// diags = append(diags, secretDiags...)


	return results,diags
	

}
