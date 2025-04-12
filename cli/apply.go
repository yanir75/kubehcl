package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
)

// type apply struct {
// 	FolderName string
// 	Name string
// }

var applydesc string = `apply will create or update existing resources managed by kubehcl
automatically searches for files with ending of .hcl`

// Apply command will validate then create the corresponding components written in the configuration files
func applyCmd() *cobra.Command {

	applyCmd := &cobra.Command{
		Use:   "apply [name] [folder]",
		Short: "Create or update resources",
		Long:  applydesc,
		Run: func(cmd *cobra.Command, args []string) {
			conf := cmd.Context().Value("settings").(*settings.EnvSettings)
			viewSettings := cmd.Context().Value("viewSettings").(*view.ViewArgs)

			client.Apply(args,conf,viewSettings)
		},
	}
	addCommonToCommand(applyCmd)

	return applyCmd

}
