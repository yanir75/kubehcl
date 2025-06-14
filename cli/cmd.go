package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Create root command for the cmd tool
func CreateRootCMD() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "kubehcl",
		Short: "Kubehcl CLI",
		Long:  "Kubehcl simplifies deployment to kubernetes using HCL",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				fmt.Println(err)
			}
		},
	}

	rootCmd.AddCommand(
		licenseCmd(),
		installCmd(),
		uninstallCmd(),
		templateCmd(),
		listCmd(),
		createCmd(),
		fmtCmd(),
		versionCmd(),
		// planCmd(),
	)
	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true

	return rootCmd
}
