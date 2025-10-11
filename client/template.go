package client

import (
	"fmt"
	"os"
	"sync"

	"github.com/hashicorp/hcl/v2"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	"kubehcl.sh/kubehcl/internal/configs"
	"kubehcl.sh/kubehcl/internal/dag"
	"kubehcl.sh/kubehcl/internal/decode"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
)

// Template expects 1 argument
// 1. Folder name which folder to decode
// Template will render the configuration and print it as json/yaml format after inserting the values
func Template(args []string, kind string, viewArguments *view.ViewArgs, cmdSettings *settings.CmdSettings) {
	folderName, diags := parseFolderArgs(args)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewArguments)
		return
	}

	varF, vars, diags := parseCmdSettings(cmdSettings)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewArguments)
		return
	}

	d, diags := configs.DecodeFolderAndModules("", folderName, "root", varF, vars, 0)
	g := &configs.Graph{
		DecodedModule: d,
	}
	diags = append(diags, g.Init()...)

	if diags.HasErrors() {
		v.DiagPrinter(diags, viewArguments)
		os.Exit(1)
	}
	var mutex sync.Mutex

	printFunc := func(v dag.Vertex) hcl.Diagnostics {
		switch resource := v.(type) {
		case *decode.DecodedResource:
			for key, value := range resource.Config {
				var resourceOutput []byte
				var err error

				if kind == "yaml" {
					resourceOutput, err = ctyyaml.Marshal(value)

				}
				if kind == "json" {
					resourceOutput, err = ctyjson.Marshal(value, value.Type())
				}

				if err != nil {
					mutex.Lock()
					defer mutex.Unlock()
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Resource couldn't be marshalled",
						Detail:   fmt.Sprintf("Resource: %s couldn't be marshalled", resource.Name),
						Subject:  &resource.DeclRange,
					})
				} else {
					fmt.Printf("# Resource: %s\n\n", key)
					fmt.Printf("%s______________________________________________\n", string(resourceOutput))
				}
			}
		}
		return nil
	}

	diags = append(diags, g.Walk(printFunc)...)

	v.DiagPrinter(diags, viewArguments)

}
