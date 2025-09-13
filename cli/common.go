package cli

import (
	"context"

	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/settings"
)

type key string

func (c key) String() string {
	return string(c)
}

const (
	settingsKey = key("settings")
	viewKey     = key("viewSettings")
)

// Adds he common flags to the command
// Example of common flag is --namespace
func addCommonToCommand(cmd *cobra.Command) {
	definitions := settings.NewSettings()
	definitions.AddFlags(cmd.PersistentFlags())

	viewSettings := settings.NewView()
	settings.AddViewFlags(viewSettings, cmd.PersistentFlags())
	ctx := context.WithValue(context.Background(), settingsKey, definitions)
	ctx = context.WithValue(ctx, viewKey, viewSettings)

	cmd.SetContext(ctx)

}

func addView(cmd *cobra.Command) {
	viewSettings := settings.NewView()
	settings.AddViewFlags(viewSettings, cmd.PersistentFlags())
	ctx := context.WithValue(context.Background(), viewKey, viewSettings)

	cmd.SetContext(ctx)
}
