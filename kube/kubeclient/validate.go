package kubeclient

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	"kubehcl.sh/kubehcl/internal/decode"
	"kubehcl.sh/kubehcl/kube/syntaxvalidator"
)

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
		if diags.HasErrors() {
			return diags
		}

		factory, validatorDiags := syntaxvalidator.New(cfg.Version)
		diags = append(diags, validatorDiags...)
		if diags.HasErrors() {
			return hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  "Couldn't build validator",
				Detail:   fmt.Sprintf("Kubernetes syntax won't be validated %s", diags.Errs()[0]),
			}}
		}
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
