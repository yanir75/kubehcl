package kubeclient

import (
	"github.com/hashicorp/hcl/v2"
	"helm.sh/helm/v4/pkg/kube"
	"kubehcl.sh/kubehcl/internal/decode"
)

// Not implemented
func (cfg *Config) Plan(resource *decode.DecodedResource) (kube.ResourceList, kube.ResourceList, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var wantedList, currentList kube.ResourceList
	// fmt.Printf("%s:%d\n",resource.Name,len(resource.Config))
	for key, value := range resource.Config {
		wanted, buildDiags := cfg.buildResource(key, value, &resource.DeclRange)
		wantedList = append(wantedList, wanted...)
		diags = append(diags, buildDiags...)
		current, buildDiags := cfg.Storage.BuildResourceFromState(wanted, key)
		currentList = append(currentList, current...)
		diags = append(diags, buildDiags...)
	}

	// if len(wantedList) == 0 {
	// 	panic("err")
	// }
	return currentList, wantedList, diags
}
