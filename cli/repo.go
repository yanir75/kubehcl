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
		Long:  "repo provides you the option to add a repository containing modules, remove a repository or list all repositories",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				fmt.Println(err)
			}
		},
	}
	repoCmd.AddCommand(
		repoAddCmd(),
		repoRemoveCmd(),
		repoListCmd(),
	)

	return repoCmd
}
