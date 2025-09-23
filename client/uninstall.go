package client

import (
	"fmt"
	"os"
	"slices"

	"github.com/hashicorp/hcl/v2"

	"kubehcl.sh/kubehcl/internal/configs"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/kube/kubeclient"
	"kubehcl.sh/kubehcl/settings"
)

// Parses arguments for uninstall command
func parseUninstallArgs(args []string) (string,string, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	if len(args) > 2 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Too many arguments required arguments are: name",
		})
		return "","", diags
	}

	if len(args) < 2 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Required arguments are :[name, folder name]",
		})
		return "","", diags
	}

	return args[0],args[1], diags

}

// Uninstall expects 1 argument
// 1. Release name, name of the release to be saved.
// The rest is environment variables and flags of the settings for example namespace otherwise it will use the default settings
// Uninstall will uninstall all resources registered to the given namespace and release name
func Uninstall(args []string, conf *settings.EnvSettings, viewArguments *view.ViewArgs,cmdSettings *settings.CmdSettings) {
	name,folderName, diags := parseUninstallArgs(args)
	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		return
	}

	varsF, vals, diags := parseCmdSettings(cmdSettings)
	d, decodeDiags := configs.DecodeFolderAndModules(name, folderName, "root", varsF, vals, 0)
	diags = append(diags, decodeDiags...)
	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		os.Exit(1)
	}

	cfg, cfgDiags := kubeclient.New(name, conf,d.BackendStorage.Kind)
	diags = append(diags, cfgDiags...)

	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		return
	}

	secrets, secretDiags := cfg.List()
	diags = append(diags, secretDiags...)

	if secretDiags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		return
	}

	if !slices.Contains(secrets, "kubehcl."+cfg.Name) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Release does not exist",
			Detail:   fmt.Sprintf("The release you provided \"%s\" does not exist in the given namespace \"%s\"", cfg.Name, conf.Namespace()),
		})
	}
	
	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		os.Exit(1)
	}

	if d.BackendStorage.Kind == "stateless" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary: "Storage kind is stateless no resource will be deleted",
			Subject: &d.BackendStorage.DeclRange,
		})
		diags = append(diags, cfg.Storage.DeleteState()...)
		if diags.HasErrors() {
			view.DiagPrinter(diags, viewArguments)
			os.Exit(1)
		}
		view.DiagPrinter(diags, viewArguments)
		os.Exit(0)
	}

	_, deleteDiags := cfg.DeleteAllResources()
	diags = append(diags, deleteDiags...)
	view.DiagPrinter(diags, viewArguments)

}
