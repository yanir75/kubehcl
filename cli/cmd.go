package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

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
		applyCmd(),
		destroyCmd(),
		templateCmd(),
		listCmd(),
	)
	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true

	return rootCmd
}
