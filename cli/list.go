package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
	"kubehcl.sh/kubehcl/settings"
)

// type apply struct {
// 	FolderName string
// 	Name string
// }

var listDesc string = `list will return all releases applied through kubehcl`

func listCmd() *cobra.Command {
	// var a apply

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list all modules",
		Long:  listDesc,
		Run: func(cmd *cobra.Command, args []string) {
			conf := cmd.Context().Value("settings").(*settings.EnvSettings)
			client.List(conf)
		},
	}
	addCommonToCommand(listCmd)


	return listCmd

}
