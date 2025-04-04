package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
)

// type apply struct {
// 	FolderName string
// 	Name string
// }

var applydesc string = `apply will create or update existing resources managed by kubehcl
automatically searches for files with ending of .hcl`

func applyCmd() *cobra.Command {
	// var a apply

	applyCmd := &cobra.Command{
		Use:   "apply [name] [folder]",
		Short: "Create or update resources",
		Long:  applydesc,
		Run: func(cmd *cobra.Command, args []string) {
			client.Apply(args)
		},
	}

	return applyCmd

}
