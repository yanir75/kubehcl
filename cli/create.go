package cli

import (
	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
	"kubehcl.sh/kubehcl/internal/view"
)

// Template prints the template which will be applied in yaml form after being rendered
func createCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create will create an example folder for you to build your inital release",
		Long:  "Create command will create multiple files to show you how to create an example configuration, this will allow a better start than writing from zero",
		Run: func(cmd *cobra.Command, args []string) {

			viewSettings := cmd.Parent().Context().Value(viewKey).(*view.ViewArgs)

			client.Create(args, viewSettings)
		},
	}

	// addCommonToCommand(createCmd)

	return createCmd

}
