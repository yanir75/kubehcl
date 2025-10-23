package cli

// Current maintainer Yanir75
import (
	"fmt"

	"github.com/spf13/cobra"
)

// Template prints the template which will be applied in yaml form after being rendered
var version = "v0.3.2"

func versionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Shows the current version of the tool",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Kubehcl %s\n", version)
		},
	}

	return versionCmd

}
