package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
	"kubehcl.sh/kubehcl/settings"
)

func destroyCmd() *cobra.Command {
	destroyCmd := &cobra.Command{
		Use:   "destroy",
		Short: "Destory all resources managed by kubehcl",
		Long:  "Destroy will destroy existing resources managed by kubehcl",
		Run: func(cmd *cobra.Command, args []string) {
			conf := cmd.Context().Value("settings").(*settings.EnvSettings)

			client.Destroy(args,conf)
		},
	}
	addCommonToCommand(destroyCmd)

	return destroyCmd

}
