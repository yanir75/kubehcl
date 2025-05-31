package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
)

// TODO: Use the same options as create and destroy in opposite option
// Uninstall will destroy the corresponding components of the given apply name
func uninstallCmd() *cobra.Command {
	destroyCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall all resources managed by kubehcl",
		Long:  "Uninstall will destroy existing resources managed by kubehcl",
		Run: func(cmd *cobra.Command, args []string) {
			conf := cmd.Context().Value(settingsKey).(*settings.EnvSettings)
			viewSettings := cmd.Context().Value(viewKey).(*view.ViewArgs)

			client.Uninstall(args, conf, viewSettings)
		},
	}
	addCommonToCommand(destroyCmd)

	return destroyCmd

}
