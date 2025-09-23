package client

import (
	"fmt"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/kube/kubeclient"
	"kubehcl.sh/kubehcl/settings"
)

// list the installations of kubehcl in the namespace
func List(conf *settings.EnvSettings, viewArguments *view.ViewArgs,storageKind string) {
	cfg, diags := kubeclient.New("", conf,storageKind)
	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
	} else {
		if secrets, diags := cfg.List(); diags.HasErrors() {
			view.DiagPrinter(diags, viewArguments)
		} else {
			for _, secret := range secrets {
				fmt.Printf("Installation: %s\n", secret)
			}
		}
	}

}
