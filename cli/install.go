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
			conf := cmd.Parent().Context().Value(settingsKey).(*settings.EnvSettings)
			viewSettings := cmd.Context().Value(viewKey).(*view.ViewArgs)
			cmdSettings := cmd.Context().Value(cmdSettingsKey).(*settings.CmdSettings)

			client.Install(args, conf, viewSettings, cmdSettings, i.CreateNamespace)
		},
	}
	// addCommonToCommand(installCmd)
	installCmd.Flags().BoolVar(&i.CreateNamespace, "create-namespace", false, "automatically create namespace")
	addView(installCmd)
	AddCmdSettings(installCmd)

	return installCmd

}
