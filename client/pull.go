package client

import (

	"github.com/hashicorp/hcl/v2"
	"kubehcl.sh/kubehcl/internal/configs"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"

)

func parsePullArgs(args []string) (string, string, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	if len(args) != 2 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Required arguments are :[repo, tag/name]",
		})
		return "", "", diags
	}
	return args[0], args[1], diags

}

func Pull(version string, envSettings *settings.EnvSettings, viewDef *view.ViewArgs, args []string,save bool) {
	repoName, tag, diags := parsePullArgs(args)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return 
	}

	_, diags = configs.Pull(version,envSettings.RepositoryConfig,repoName,tag,true)
	v.DiagPrinter(diags,viewDef)
}
