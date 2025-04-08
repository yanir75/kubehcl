package cli

import (
	"context"

	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/settings"
)

func addCommonToCommand(cmd *cobra.Command) {
	settings := settings.New()
	settings.AddFlags(cmd.Flags())
	ctx := context.WithValue(context.Background(),"settings",settings)
	cmd.SetContext(ctx)

}