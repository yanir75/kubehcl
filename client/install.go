package client

import (
	"fmt"
	"os"
	"sync"
	"time"

	"helm.sh/helm/v4/pkg/kube"

	"github.com/hashicorp/hcl/v2"
	"kubehcl.sh/kubehcl/internal/configs"
	"kubehcl.sh/kubehcl/internal/dag"
	"kubehcl.sh/kubehcl/internal/decode"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/kube/kubeclient"
	"kubehcl.sh/kubehcl/settings"
)

type installResult struct {
	run       bool
	name      string
	operation string
}

func printUpdateFunc(res *installResult, wg *sync.WaitGroup) {
	i := 0
	for res.run {
		fmt.Printf("Creating/Updating kube_resource: %s (%d seconds has passed) \n", res.name, i*10)
		i++
		time.Sleep(time.Second * 10)
	}
	fmt.Printf("%s kube_resource: %s\n", res.operation, res.name)
	wg.Done()

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

// Install expects 2 arguments
// 1. Release name, name of the release to be saved.
// 2. Folder name which folder to decode
// The rest is environment variables and flags of the settings for example namespace otherwise it will use the default settings
// After parsing the variables install will decode the folder, validate the configuration and create the components.
func Install(args []string, conf *settings.EnvSettings, viewArguments *view.ViewArgs, cmdSettings *settings.CmdSettings, createNamespace bool) {
	name, folderName, diags := parseInstallArgs(args)
	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		return
	}

	varsF, vals, diags := parseCmdSettings(cmdSettings)

	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		return
	}

	d, decodeDiags := configs.DecodeFolderAndModules(name, folderName, "root", varsF, vals, 0)
	diags = append(diags, decodeDiags...)
	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		os.Exit(1)
	}
	g := &configs.Graph{
		DecodedModule: d,
	}
	diags = append(diags, g.Init()...)
	cfg, cfgDiags := kubeclient.New(name, conf, d.BackendStorage.Kind)
	diags = append(diags, cfgDiags...)

	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		os.Exit(1)
	}
	view.DiagPrinter(diags, viewArguments)

	diags = cfg.VerifyInstall(createNamespace)
	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		os.Exit(1)
	}

	var results = kube.Result{}
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
	if len(validateDiags) > 0 {
		view.DiagPrinter(validateDiags[0:1], viewArguments)
		os.Exit(1)
	}

	if !diags.HasErrors() {
		diags = append(diags, g.Walk(createFunc)...)
		// if cfg.StorageKind != "stateless" {
		saved, _, delDiags := cfg.DeleteResources()
		diags = append(diags, delDiags...)
		for key := range saved {
			fmt.Printf("Deleted resource: %s\n", key)
		}
		// }
	}
	diags = append(diags, cfg.Storage.UpdateState()...)

	view.DiagPrinter(diags, viewArguments)

}
