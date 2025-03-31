package cli

import (
	"fmt"

	"kubehcl.sh/kubehcl/client"

	"github.com/spf13/cobra"
)


func CreateRootCMD() *cobra.Command{
	rootCmd := &cobra.Command{
		Use: "kubehcl",
		Short: "Kubehcl CLI",
		Long: "Kubehcl simplifies deployment to kubernetes using HCL",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err!=nil{
				fmt.Println(err)
			}
		},
	}

	applyCmd := &cobra.Command{
		Use: "apply",
		Short: "Create or update resources",
		Long: "apply will create or update existing resources managed by kubehcl",
		Run: func(cmd *cobra.Command, args []string) {
			client.Apply()
		},
	}

	destroyCmd := &cobra.Command{
		Use: "destroy",
		Short: "Destory all resources managed by kubehcl",
		Long: "Destroy will destroy existing resources managed by kubehcl",
		Run: func(cmd *cobra.Command, args []string) {
			client.Destroy()
		},
	}
	

	
	

	rootCmd.AddCommand(	
						applyCmd,
						destroyCmd,
						templateCmd(),
					)
	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true

	return rootCmd
}