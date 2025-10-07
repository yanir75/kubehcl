package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
	"kubehcl.sh/kubehcl/internal/logging"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
)

var listDesc string = `list will return all releases applied through kubehcl`

// List will list all deployments in a given namespace
func listCmd() *cobra.Command {

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list all modules",
		Long:  listDesc,
		Run: func(cmd *cobra.Command, args []string) {
			conf := cmd.Parent().Context().Value(settingsKey).(*settings.EnvSettings)
			viewSettings := cmd.Parent().Context().Value(viewKey).(*view.ViewArgs)
			logging.SetLogger(conf.Debug)

			client.List(conf, viewSettings, "kube_secret")
		},
	}
	// addCommonToCommand(listCmd)

	return listCmd

}
