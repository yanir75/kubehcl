package settings

import (
	"github.com/spf13/pflag"
)

type CmdSettings struct {
	VarsFile string
	Vars     []string
}

// Apply view settings to the diagprinter

func NewCmdSettings() *CmdSettings {
	cmdSettings := &CmdSettings{
		VarsFile: envOr("KUBEHCL_VARS", "kubehcl.tfvars"),
	}

	return cmdSettings
}

func AddCmdSettings(c *CmdSettings, fs *pflag.FlagSet) {
	fs.StringVar(&c.VarsFile, "var-file", c.VarsFile, "Vars filename to load values into variables")
	fs.StringSliceVar(&c.Vars, "var", make([]string, 0), "Set a specific variable's value must have an equal within it")
}
