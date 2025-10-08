package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Create module command for the cmd tool
func repoCmd() *cobra.Command {
	repoCmd := &cobra.Command{
		Use:   "repo",
		Short: "allows management of modules in repositories with subcommands",
		Long:  "module provides you the option to download or install modules",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				fmt.Println(err)
			}
		},
	}
	repoCmd.AddCommand(repoAddCmd())

	return repoCmd
}
