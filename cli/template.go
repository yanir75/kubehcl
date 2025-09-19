package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
)

type template struct {
	Kind      string
	Namespace string
}

// Template prints the template which will be applied in yaml form after being rendered
func templateCmd() *cobra.Command {
	var t template
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Print the resources which will be created in yaml or json format",
		Long:  "Template converts the hcl to yaml in order to view the kubernetes yamls which will be applied and created in your environment",
		Run: func(cmd *cobra.Command, args []string) {

			// conf := cmd.Context().Value(settingsKey).(*settings.EnvSettings)
			viewSettings := cmd.Context().Value(viewKey).(*view.ViewArgs)
			cmdSettings := cmd.Context().Value(cmdSettingsKey).(*settings.CmdSettings)

			switch t.Kind {
			case "yaml":
				client.Template(args, "yaml", viewSettings, cmdSettings)
			case "json":
				client.Template(args, "json", viewSettings, cmdSettings)
			default:
				fmt.Println("Valid arguments for kind are [yaml, json]")
				os.Exit(1)
			}
		},
	}

	templateCmd.Flags().StringVar(&t.Kind, "kind", "yaml", "prints the template in yaml or json format")
	// templateCmd.Flags().StringVar(&t.Namespace, "namespace", "default", "prints the template in yaml or json format")

	addView(templateCmd)
	AddCmdSettings(templateCmd)

	return templateCmd

}
