package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
	"kubehcl.sh/kubehcl/internal/logging"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
)

// Create module command for the cmd tool
func repoListCmd() *cobra.Command {

	repoListCommand := &cobra.Command{
		Use:   "list",
		Short: "lists all added repositories",
		Long:  "list allows you to see which repositories were added",
		Run: func(cmd *cobra.Command, args []string) {
			viewSettings := cmd.Parent().Parent().Context().Value(viewKey).(*view.ViewArgs)
			conf := cmd.Parent().Parent().Context().Value(settingsKey).(*settings.EnvSettings)
			logging.SetLogger(conf.Debug)
			client.ListRepos(conf,viewSettings,args)
		},
	}


	return repoListCommand
}
