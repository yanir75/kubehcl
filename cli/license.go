package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Will print the licenses related to the project
func licenseCmd() *cobra.Command {
	licenseCmd := &cobra.Command{
		Use:   "license",
		Short: "license printer",
		Long:  "Prints the licenses associated with the project",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("This project is dual license by MPL 2.0 and Apache 2.0")
			fmt.Printf("Apache 2.0: %s\nMPL 2.0: %s\n", "https://www.apache.org/licenses/LICENSE-2.0", "https://www.mozilla.org/en-US/MPL/2.0/")

		},
	}
	return licenseCmd
}
