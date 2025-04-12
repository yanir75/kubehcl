package cli

import (
	"context"

	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/settings"
)

// Adds he common flags to the command
// Example of common flag is --namespace
func addCommonToCommand(cmd *cobra.Command) {
	definitions := settings.NewSettings()
	definitions.AddFlags(cmd.Flags())

	viewSettings := settings.NewView()
	settings.AddViewFlags(viewSettings,cmd.Flags())
	ctx := context.WithValue(context.Background(),"settings",definitions)
	ctx = context.WithValue(ctx,"viewSettings",viewSettings)

	cmd.SetContext(ctx)

}