package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"kubehcl.sh/kubehcl/client"
)

type template struct {
	Kind string
}

func templateCmd() *cobra.Command {
	var t template
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Print the resources which will be created in yaml or json format",
		Long:  "Template converts the hcl to yaml in order to view the kubernetes yamls which will be applied and created in your environment",
		Run: func(_ *cobra.Command, _ []string) {
			switch t.Kind {
			case "yaml":
				client.Template("yaml")
			case "json":
				client.Template("json")
			default:
				fmt.Println("Valid arguments for kind are [yaml, json]")
				os.Exit(1)
			}
		},
	}

	templateCmd.Flags().StringVar(&t.Kind, "kind", "yaml", "prints the template in yaml or json format")

	return templateCmd

}
