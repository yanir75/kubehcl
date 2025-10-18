package client

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/afero"
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

func Pull(version string, envSettings *settings.EnvSettings, viewDef *view.ViewArgs, args []string, save bool) (afero.Fs, hcl.Diagnostics) {
	repoName, tag, diags := parsePullArgs(args)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return nil, diags
	}

	appFs, diags := configs.Pull(version, envSettings.RepositoryConfig, repoName, tag, save)
	v.DiagPrinter(diags, viewDef)
	return appFs, diags
}
