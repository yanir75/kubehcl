package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
)

// Destroy will destroy the corresponding components of the given apply name
func destroyCmd() *cobra.Command {
	destroyCmd := &cobra.Command{
		Use:   "destroy",
		Short: "Destory all resources managed by kubehcl",
		Long:  "Destroy will destroy existing resources managed by kubehcl",
		Run: func(cmd *cobra.Command, args []string) {
			conf := cmd.Context().Value(settingsKey).(*settings.EnvSettings)
			viewSettings := cmd.Context().Value(viewKey).(*view.ViewArgs)

			client.Destroy(args, conf, viewSettings)
		},
	}
	addCommonToCommand(destroyCmd)

	return destroyCmd

}
