package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Template prints the template which will be applied in yaml form after being rendered
var version = "v0.1.12"

func versionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Shows the current version of the tool",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Kuhbecl %s", version)
		},
	}

	return versionCmd

}
