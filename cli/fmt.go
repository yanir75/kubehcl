package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
	"kubehcl.sh/kubehcl/internal/view"
)

// type install struct {
// 	FolderName string
// 	Name string
// }

var fmtdesc string = `fmt will format all files in the folder`

type fmtConf struct {
	recursive bool
}
// Apply command will validate then create the corresponding components written in the configuration files
func fmtCmd() *cobra.Command {
	var f fmtConf

	fmtCmd := &cobra.Command{
		Use:   "fmt [folder]",
		Short: "format all files in the folder",
		Long:  fmtdesc,
		Run: func(cmd *cobra.Command, args []string) {
			viewSettings := cmd.Context().Value(viewKey).(*view.ViewArgs)
			client.Fmt(args,viewSettings,f.recursive)
		},
	}
	fmtCmd.Flags().BoolVar(&f.recursive, "recursive", false, "prints the template in yaml or json format")
	addCommonToCommand(fmtCmd)

	return fmtCmd

}
