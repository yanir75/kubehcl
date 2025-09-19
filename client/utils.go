package client

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"kubehcl.sh/kubehcl/settings"
)

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
