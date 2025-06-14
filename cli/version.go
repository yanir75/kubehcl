package cli

import (


	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
	"kubehcl.sh/kubehcl/internal/view"
)



// Template prints the template which will be applied in yaml form after being rendered
func versionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Shows the current version of the tool",
		Run: func(cmd *cobra.Command, args []string) {

			viewSettings := cmd.Context().Value(viewKey).(*view.ViewArgs)

			client.Create(args,viewSettings)
		},
	}

	return versionCmd

}
