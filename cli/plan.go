package cli

// import (
// 	"github.com/spf13/cobra"
// 	"kubehcl.sh/kubehcl/client"
// 	"kubehcl.sh/kubehcl/internal/view"
// 	"kubehcl.sh/kubehcl/settings"
// )

// var plandesc string = `plan will show the difference betwee nexisting resources managed by kubehcl and wanted configuration
// automatically searches for files with ending of .hcl`

// func planCmd() *cobra.Command {
// 	// var a apply

// 	planCmd := &cobra.Command{
// 		Use:   "plan [name] [folder]",
// 		Short: "Show changes that will be made",
// 		Long:  plandesc,
// 		Run: func(cmd *cobra.Command, args []string) {
// 			conf := cmd.Context().Value(settingsKey).(*settings.EnvSettings)
// 			viewSettings := cmd.Context().Value(viewKey).(*view.ViewArgs)
// 			client.Plan(args,conf,viewSettings)
// 		},
// 	}

// 	addCommonToCommand(planCmd)

// 	return planCmd

// }
