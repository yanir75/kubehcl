package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
	"kubehcl.sh/kubehcl/internal/logging"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
)

// Create module command for the cmd tool
func repoRemoveCmd() *cobra.Command {

	repoRemoveCommand := &cobra.Command{
		Use:   "remove [NAME]",
		Short: "add a repository of modules",
		Long:  "add provides you the option to add a repository which contains modules modules",
		Run: func(cmd *cobra.Command, args []string) {
			viewSettings := cmd.Parent().Parent().Context().Value(viewKey).(*view.ViewArgs)
			conf := cmd.Parent().Parent().Context().Value(settingsKey).(*settings.EnvSettings)
			logging.SetLogger(conf.Debug)
			client.RemoveRepo(conf,viewSettings,args)
		},
	}


	return repoRemoveCommand
}
