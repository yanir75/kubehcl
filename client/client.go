package client

import (
	"fmt"
	"os"
	"sync"

	"helm.sh/helm/v4/pkg/kube"

	"github.com/hashicorp/hcl/v2"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	"kubehcl.sh/kubehcl/internal/configs"
	"kubehcl.sh/kubehcl/internal/dag"
	"kubehcl.sh/kubehcl/internal/decode"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/kube/kubeclient"
)

func Apply(args []string) {
	name, folderName, diags := parseApplyArgs(args)
	if diags.HasErrors() {
		view.DiagPrinter(diags)
		return
	}

	d, decodeDiags := configs.DecodeFolderAndModules(folderName, "root", 0)
	diags = append(diags, decodeDiags...)
	g := &configs.Graph{
		DecodedModule: d,
	}
	diags = append(diags, g.Init()...)
	cfg, cfgDiags := kubeclient.New(name)
	diags = append(diags, cfgDiags...)

	if diags.HasErrors() {
		view.DiagPrinter(diags)
		os.Exit(1)
	}

	var results *kube.Result = &kube.Result{}
	var mutex sync.Mutex
	validateFunc := func(v dag.Vertex) hcl.Diagnostics {
		switch tt := v.(type) {
		case *decode.DecodedResource:
			return cfg.Validate(tt)
		}
		return nil
	}
	createFunc := func(v dag.Vertex) hcl.Diagnostics {
		switch tt := v.(type) {
		case *decode.DecodedResource:
			fmt.Printf("Creating/Updating resource :%s\n", tt.Name)
			res, createDiags := cfg.Create(tt)
			if !createDiags.HasErrors() {
				mutex.Lock()
				defer mutex.Unlock()
				results.Created = append(results.Created, res.Created...)
				results.Updated = append(results.Updated, res.Updated...)
				results.Deleted = append(results.Deleted, res.Deleted...)
				fmt.Printf("Created/Updated resource :%s\n", tt.Name)
			}
			return createDiags
			// fmt.Printf("%s",asdf.Created[0])
		}
		return nil
	}

	diags = append(diags, g.Walk(validateFunc)...)
	if !diags.HasErrors() {
		diags = append(diags, g.Walk(createFunc)...)
		cfg.DeleteResources()
	}
	diags = append(diags, cfg.UpdateSecret()...)

	view.DiagPrinter(diags)

}

func Template(kind string) {
	d, diags := configs.DecodeFolderAndModules(".", "root", 0)
	g := &configs.Graph{
		DecodedModule: d,
	}
	diags = append(diags, g.Init()...)

	if diags.HasErrors() {
		view.DiagPrinter(diags)
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
					fmt.Printf("Resource: %s\n\n", key)
					fmt.Printf("%s______________________________________________\n", string(resourceOutput))
				}
			}
		}
		return nil
	}

	diags = append(diags, g.Walk(printFunc)...)

	view.DiagPrinter(diags)

}

func Destroy(args []string) {
	name, diags := parseDestroyArgs(args)
	if diags.HasErrors() {
		view.DiagPrinter(diags)
		return
	}
	cfg, cfgDiags := kubeclient.New(name)
	diags = append(diags, cfgDiags...)

	if diags.HasErrors() {
		view.DiagPrinter(diags)
	} else {
		cfg.DeleteAllResources()
	}
}

func List() {
	cfg, diags := kubeclient.New("")
	if diags.HasErrors() {
		view.DiagPrinter(diags)
	} else {
		if secrets, diags := cfg.List(); diags.HasErrors() {
			view.DiagPrinter(diags)
		} else {
			for _, secret := range secrets {
				fmt.Printf("module: %s\n", secret)
			}
		}
	}

}

func parseDestroyArgs(args []string) (string, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	if len(args) > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Too many arguments required arguments are: name",
		})
		return "", diags
	}

	if len(args) < 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Too few arguments required arguments are: name",
		})
		return "", diags
	}

	return args[0], diags

}

func parseApplyArgs(args []string) (string, string, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	if len(args) != 2 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Required arguments are :[name, folder name]",
		})
		return "", "", diags
	}
	return args[0], args[1], diags

}
