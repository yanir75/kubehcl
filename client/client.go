package client

import (
	"fmt"
	"os"
	"sync"
	"time"

	"helm.sh/helm/v4/pkg/kube"

	"slices"

	"github.com/hashicorp/hcl/v2"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	"kubehcl.sh/kubehcl/internal/configs"
	"kubehcl.sh/kubehcl/internal/dag"
	"kubehcl.sh/kubehcl/internal/decode"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/kube/kubeclient"
	"kubehcl.sh/kubehcl/settings"
)

// Install expects 2 arguments
// 1. Release name, name of the release to be saved.
// 2. Folder name which folder to decode
// The rest is environment variables and flags of the settings for example namespace otherwise it will use the default settings
// After parsing the variables install will decode the folder, validate the configuration and create the components.
type installResult struct {
	run       bool
	name      string
	operation string
}

func printUpdateFunc(res *installResult, wg *sync.WaitGroup) {
	i := 0
	for res.run {
		fmt.Printf("Creating/Updating resource: %s (%d seconds has passed) \n", res.name, i*10)
		i++
		time.Sleep(time.Second * 10)
	}
	fmt.Printf("%s resource: %s\n", res.operation, res.name)
	wg.Done()

}

func Install(args []string, conf *settings.EnvSettings, viewArguments *view.ViewArgs) {
	name, folderName, diags := parseInstallArgs(args)
	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		return
	}

	d, decodeDiags := configs.DecodeFolderAndModules(folderName, "root", 0, conf.Namespace())
	diags = append(diags, decodeDiags...)
	g := &configs.Graph{
		DecodedModule: d,
	}
	diags = append(diags, g.Init()...)
	cfg, cfgDiags := kubeclient.New(name, conf)
	diags = append(diags, cfgDiags...)

	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
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
			// fmt.Printf("Creating/Updating resource: %s\n", tt.Name)
			installRes := &installResult{}
			installRes.name = tt.Name
			installRes.run = true
			installRes.operation = "Created"
			var wg sync.WaitGroup
			wg.Add(1)
			go printUpdateFunc(installRes, &wg)
			res, createDiags := cfg.Create(tt)
			if !createDiags.HasErrors() {
				if len(res.Updated) > 0 {
					installRes.operation = "Updated"
				}
				if len(res.Deleted) > 0 {
					installRes.operation = "Deleted"
				}
				mutex.Lock()
				defer mutex.Unlock()
				results.Created = append(results.Created, res.Created...)
				results.Updated = append(results.Updated, res.Updated...)
				results.Deleted = append(results.Deleted, res.Deleted...)
				// fmt.Printf("%s resource: %s\n",operation, tt.Name)
			} else {
				installRes.operation = "Failed to perform any action on"
			}
			installRes.run = false
			wg.Wait()

			return createDiags
			// fmt.Printf("%s",asdf.Created[0])
		}
		return nil
	}
	validateDiags := g.Walk(validateFunc)
	if len(validateDiags) > 1 {
		view.DiagPrinter(validateDiags[0:1], viewArguments)
	}

	if !diags.HasErrors() {
		diags = append(diags, g.Walk(createFunc)...)
		saved, _, delDiags := cfg.DeleteResources()
		diags = append(diags, delDiags...)
		for key := range saved {
			fmt.Printf("Deleted resource: %s\n", key)
		}
	}
	diags = append(diags, cfg.UpdateSecret()...)

	view.DiagPrinter(diags, viewArguments)

}

// Template expects 1 argument
// 1. Folder name which folder to decode
// Template will render the configuration and print it as json/yaml format after inserting the values
func Template(args []string, kind string, namespace string, viewArguments *view.ViewArgs) {
	folderName, diags := parseTemplateArgs(args)
	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		return
	}

	d, diags := configs.DecodeFolderAndModules(folderName, "root", 0, namespace)
	g := &configs.Graph{
		DecodedModule: d,
	}
	diags = append(diags, g.Init()...)

	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
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

	view.DiagPrinter(diags, viewArguments)

}

// Uninstall expects 1 argument
// 1. Release name, name of the release to be saved.
// The rest is environment variables and flags of the settings for example namespace otherwise it will use the default settings
// Uninstall will uninstall all resources registered to the given namespace and release name
func Uninstall(args []string, conf *settings.EnvSettings, viewArguments *view.ViewArgs) {
	name, diags := parseUninstallArgs(args)
	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		return
	}
	cfg, cfgDiags := kubeclient.New(name, conf)
	diags = append(diags, cfgDiags...)

	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		return
	}

	secrets, secretDiags := cfg.List()
	diags = append(diags, secretDiags...)

	if secretDiags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		return
	}

	if !slices.Contains(secrets, "kubehcl."+cfg.Name) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Release does not exist",
			Detail:   fmt.Sprintf("The release you provided \"%s\" does not exist in the given namespace \"%s\"", cfg.Name, conf.Namespace()),
		})
	}
	_,deleteDiags := cfg.DeleteAllResources()
	diags = append(diags, deleteDiags...)
	view.DiagPrinter(diags, viewArguments)

}

func List(conf *settings.EnvSettings, viewArguments *view.ViewArgs) {
	cfg, diags := kubeclient.New("", conf)
	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
	} else {
		if secrets, diags := cfg.List(); diags.HasErrors() {
			view.DiagPrinter(diags, viewArguments)
		} else {
			for _, secret := range secrets {
				fmt.Printf("module: %s\n", secret)
			}
		}
	}

}

// Parses arguments for template command
func parseTemplateArgs(args []string) (string, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	if len(args) > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Too many arguments required arguments are: folder",
		})
		return "", diags
	}

	if len(args) < 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Insufficient number of arguments required arguments are: folder",
		})
		return "", diags
	}

	return args[0], diags
}

// Parses arguments for uninstall command
func parseUninstallArgs(args []string) (string, hcl.Diagnostics) {
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
			Summary:  "Insufficient number of arguments required arguments are: name",
		})
		return "", diags
	}

	return args[0], diags

}

// Parses arguemtns for install command
func parseInstallArgs(args []string) (string, string, hcl.Diagnostics) {
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

func Create(args []string, viewArguments *view.ViewArgs) {
	name, diags := parseTemplateArgs(args)
	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		return
	}
	cacheDir(name)
}

// func Plan(args []string,conf *settings.EnvSettings,viewArguments *view.ViewArgs) {
// 	name, folderName, diags := parseInstallArgs(args)
// 	if diags.HasErrors() {
// 		view.DiagPrinter(diags,viewArguments)
// 		return
// 	}

// 	d, decodeDiags := configs.DecodeFolderAndModules(folderName, "root", 0,conf.Namespace())
// 	diags = append(diags, decodeDiags...)
// 	g := &configs.Graph{
// 		DecodedModule: d,
// 	}
// 	diags = append(diags, g.Init()...)
// 	cfg, cfgDiags := kubeclient.New(name,conf)
// 	diags = append(diags, cfgDiags...)

// 	if diags.HasErrors() {
// 		view.DiagPrinter(diags,viewArguments)
// 		os.Exit(1)
// 	}

// 	var mutex sync.Mutex
// 	validateFunc := func(v dag.Vertex) hcl.Diagnostics {
// 		switch tt := v.(type) {
// 		case *decode.DecodedResource:
// 			return cfg.Validate(tt)
// 		}
// 		return nil
// 	}
// 	wantedMap := make(map[string]kube.ResourceList)
// 	currentMap := make(map[string]kube.ResourceList)

// 	planFunc := func(v dag.Vertex) hcl.Diagnostics {
// 		switch tt := v.(type) {
// 		case *decode.DecodedResource:
// 			if len(tt.Config) > 0 {
// 				// fmt.Printf("%s\n",tt.Name)
// 				current,wanted, planDiags := cfg.Plan(tt)
// 				if !planDiags.HasErrors() {
// 					mutex.Lock()
// 					defer mutex.Unlock()
// 					wantedMap[tt.Name] = wanted
// 					currentMap[tt.Name] = current
// 				}
// 				return planDiags
// 			}
// 			// fmt.Printf("%s",asdf.Created[0])
// 		}
// 		return nil
// 	}

// 	diags = append(diags, g.Walk(validateFunc)...)
// 	if !diags.HasErrors() {
// 		diags = append(diags, g.Walk(planFunc)...)
// 		cfg.DeleteResources()
// 	}
// 	if diags.HasErrors(){
// 		view.DiagPrinter(diags,viewArguments)
// 		return
// 	}
// 	// diags = append(diags, cfg.UpdateSecret()...)

// 	view.PlanPrinter(wantedMap,currentMap)

// }

func Fmt(args []string, viewArguments *view.ViewArgs, recurisve bool) {
	folder, diags := parseTemplateArgs(args)
	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		return
	}
	diags = fmtDir(folder, recurisve)

	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		return
	}
}
