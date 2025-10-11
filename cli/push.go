package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
	"kubehcl.sh/kubehcl/internal/logging"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
)

// Create module command for the cmd tool
func pushCmd() *cobra.Command {
	PushCmd := &cobra.Command{
		Use:   "push [Folder] [REPO NAME] [TAG]",
		Short: "push a module from the OCI, push can only push to OCI",
		Long:  "push a moduel from the OCI with the given tag and name, this will push the folder to the given repo",
		Run: func(cmd *cobra.Command, args []string) {
			viewSettings := cmd.Parent().Context().Value(viewKey).(*view.ViewArgs)
			conf := cmd.Parent().Context().Value(settingsKey).(*settings.EnvSettings)
			logging.SetLogger(conf.Debug)
			client.Push(conf, viewSettings, args)
		},
	}

	return PushCmd
}
