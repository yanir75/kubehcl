package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
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
			conf := cmd.Context().Value(settingsKey).(*settings.EnvSettings)
			viewSettings := cmd.Context().Value(viewKey).(*view.ViewArgs)

			client.List(conf, viewSettings)
		},
	}
	// addCommonToCommand(listCmd)

	return listCmd

}
