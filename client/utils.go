package client

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"kubehcl.sh/kubehcl/internal/terminal"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
)

var v *view.View = view.NewView(&terminal.Streams{
	Stdout: &terminal.OutputStream{
		File: os.Stdout,
	},
	Stderr: &terminal.OutputStream{
		File: os.Stderr,
	},
	Stdin: &terminal.InputStream{
		File: os.Stdin,
	},
})

func parseCmdSettings(c *settings.CmdSettings) (string, []string, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	for _, val := range c.Vars {
		if i := strings.Index(val, "="); i == -1 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Variable must have '=' sign in it",
				Detail:   fmt.Sprintf("Variable '%s' does not have '=' sign in it", val),
			})
		}
	}

	return c.VarsFile, c.Vars, diags
}
