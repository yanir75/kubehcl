package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
)

func applyCmd() *cobra.Command {
	applyCmd := &cobra.Command{
		Use: "apply",
		Short: "Create or update resources",
		Long: "apply will create or update existing resources managed by kubehcl",
		Run: func(cmd *cobra.Command, args []string) {
			client.Apply()
		},
	}

	return applyCmd

}