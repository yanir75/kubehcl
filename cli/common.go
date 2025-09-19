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
	settingsKey    = key("settings")
	viewKey        = key("viewSettings")
	cmdSettingsKey = key("cmdSettings")
)

// Adds he common flags to the command
// Example of common flag is --namespace
func addCommonToCommand(cmd *cobra.Command) {

	definitions := settings.NewSettings()
	definitions.AddFlags(cmd.PersistentFlags())

	ctx := context.WithValue(context.Background(), settingsKey, definitions)

	cmd.SetContext(ctx)

}

func addView(cmd *cobra.Command) {
	viewSettings := settings.NewView()
	settings.AddViewFlags(viewSettings, cmd.PersistentFlags())
	ctx := context.WithValue(context.Background(), viewKey, viewSettings)
	cmd.SetContext(ctx)
}

func AddCmdSettings(cmd *cobra.Command) {
	cmdSettings := settings.NewCmdSettings()
	settings.AddCmdSettings(cmdSettings, cmd.Flags())
	ctx := context.WithValue(cmd.Context(), cmdSettingsKey, cmdSettings)

	cmd.SetContext(ctx)
}
