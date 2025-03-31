package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
)

func destroyCmd() *cobra.Command {
	destroyCmd := &cobra.Command{
		Use: "destroy",
		Short: "Destory all resources managed by kubehcl",
		Long: "Destroy will destroy existing resources managed by kubehcl",
		Run: func(cmd *cobra.Command, args []string) {
			client.Destroy(args)
		},
	}

	return destroyCmd

}