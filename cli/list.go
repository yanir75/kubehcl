package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
)

// type apply struct {
// 	FolderName string
// 	Name string
// }


var listDesc string = `list will return all releases applied through kubehcl`

func listCmd() *cobra.Command {
	// var a apply 

	applyCmd := &cobra.Command{
		Use: "list",
		Short: "list all modules",
		Long: listDesc,
		Run: func(cmd *cobra.Command, args []string) {
			client.List()
		},
	}

	return applyCmd

}
