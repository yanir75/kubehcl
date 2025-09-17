package client

import (
	"fmt"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/kube/kubeclient"
	"kubehcl.sh/kubehcl/settings"
)

func List(conf *settings.EnvSettings, viewArguments *view.ViewArgs) {
	cfg, diags := kubeclient.New("", conf)
	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
	} else {
		if secrets, diags := cfg.List(); diags.HasErrors() {
			view.DiagPrinter(diags, viewArguments)
		} else {
			for _, secret := range secrets {
				fmt.Printf("module: %s\n", secret)
			}
		}
	}

}