package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
)

// type install struct {
// 	FolderName string
// 	Name string
// }

var installdesc string = `install will create or update existing resources managed by kubehcl
automatically searches for files with ending of .hcl`


type install struct {
	CreateNamespace bool
}
// Apply command will validate then create the corresponding components written in the configuration files
func installCmd() *cobra.Command {

	var i install

	installCmd := &cobra.Command{
		Use:   "install [name] [folder]",
		Short: "Create or update resources",
		Long:  installdesc,
		Run: func(cmd *cobra.Command, args []string) {
			conf := cmd.Context().Value(settingsKey).(*settings.EnvSettings)
			viewSettings := cmd.Context().Value(viewKey).(*view.ViewArgs)
			
			client.Install(args, conf, viewSettings,i.CreateNamespace)
		},
	}
	// addCommonToCommand(installCmd)
	installCmd.Flags().BoolVar(&i.CreateNamespace, "create-namespace", false, "automatically create namespace")

	return installCmd

}
