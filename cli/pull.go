package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
	"kubehcl.sh/kubehcl/internal/logging"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
)

type pull struct {
	Version string
}

// Create module command for the cmd tool
func pullCmd() *cobra.Command {
	p := &pull{}
	PullCmd := &cobra.Command{
		Use:   "pull [REPO NAME] [TAG/NAME]",
		Short: "pulls a module from the repo",
		Long:  "pulls a moduel from the repo with the given tag and name, this will createa folder. Module must be tar.gz",
		Run: func(cmd *cobra.Command, args []string) {
			viewSettings := cmd.Parent().Context().Value(viewKey).(*view.ViewArgs)
			conf := cmd.Parent().Context().Value(settingsKey).(*settings.EnvSettings)
			logging.SetLogger(conf.Debug)
			_,_ = client.Pull(p.Version, conf, viewSettings, args, true)
		},
	}

	PullCmd.Flags().StringVar(&p.Version, "version", "", "version of the module to pull, version can only be used in repos")

	return PullCmd
}
