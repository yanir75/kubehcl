package settings

import (
	"github.com/spf13/pflag"
	"kubehcl.sh/kubehcl/internal/view"
)

// Apply view settings to the diagprinter


func NewView() *view.ViewArgs {
	view := &view.ViewArgs{
		NoColor: envBoolOr("KUBEHCL_NOCOLOR", false),
	}


	return view
}

func AddViewFlags(v *view.ViewArgs, fs *pflag.FlagSet) {
	fs.BoolVar(&v.NoColor, "no-color", v.NoColor, "No color will not print the colors of the output in any command")
}