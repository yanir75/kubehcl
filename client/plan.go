package client

import (
	"os"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"helm.sh/helm/v4/pkg/kube"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kubehcl.sh/kubehcl/internal/configs"
	"kubehcl.sh/kubehcl/internal/dag"
	"kubehcl.sh/kubehcl/internal/decode"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/kube/kubeclient"
	"kubehcl.sh/kubehcl/settings"
)

func removeUnnecessaryFields(m map[string]interface{}) {
	delete(m, "status")

	meta := m["metadata"].(map[string]interface{})
	delete(meta, "uid")
	delete(meta, "creationTimestamp")
	delete(meta, "resourceVersion")
	delete(meta, "generation")
	delete(meta, "selfLink")

	delete(meta, "managedFields")
}

func adjustCmp(m map[string]*view.CompareResources) {
	for _, value := range m {
		if value.Current != nil {
			cur := value.Current.(*unstructured.Unstructured)
			removeUnnecessaryFields(cur.Object)
		}
		if value.Wanted != nil {
			wanted := value.Wanted.(*unstructured.Unstructured)
			removeUnnecessaryFields(wanted.Object)
		}
	}
}

func Plan(args []string, conf *settings.EnvSettings, viewArguments *view.ViewArgs, cmdSettings *settings.CmdSettings) {
	name, folderName, diags := parseInstallArgs(args)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewArguments)
		return
	}

	varsF, vals, diags := parseCmdSettings(cmdSettings)

	if diags.HasErrors() {
		v.DiagPrinter(diags, viewArguments)
		return
	}

	d, decodeDiags := configs.DecodeFolderAndModules(name, folderName, "root", varsF, vals, 0)
	diags = append(diags, decodeDiags...)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewArguments)
		return
	}

	if d.BackendStorage.Kind == "stateless" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Can't use plan in stateless mode",
		})
	}
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewArguments)
		return
	}

	g := &configs.Graph{
		DecodedModule: d,
	}
	diags = append(diags, g.Init()...)
	cfg, cfgDiags := kubeclient.New(name, conf, d.BackendStorage.Kind)
	diags = append(diags, cfgDiags...)

	if diags.HasErrors() {
		v.DiagPrinter(diags, viewArguments)
		os.Exit(1)
	}

	var mutex sync.Mutex
	validateFunc := func(v dag.Vertex) hcl.Diagnostics {
		switch tt := v.(type) {
		case *decode.DecodedResource:
			return cfg.Validate(tt)
		}
		return nil
	}
	wantedMap := make(map[string]kube.ResourceList)

	planFunc := func(v dag.Vertex) hcl.Diagnostics {
		switch tt := v.(type) {
		case *decode.DecodedResource:
			if len(tt.Config) > 0 {
				// fmt.Printf("%s\n",tt.Name)
				wanted, planDiags := cfg.BuildResource(tt)
				if !planDiags.HasErrors() {
					mutex.Lock()
					defer mutex.Unlock()
					for key, value := range wanted {
						wantedMap[key] = value
					}
				}
				return planDiags
			}
			// fmt.Printf("%s",asdf.Created[0])
		}
		return nil
	}

	diags = append(diags, g.Walk(validateFunc)...)
	if !diags.HasErrors() {
		diags = append(diags, g.Walk(planFunc)...)
	}
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewArguments)
		return
	}

	currentMap, diags := cfg.GetStateResourcesCurrentState()
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewArguments)
		return
	}
	cmps, diags := cfg.CompareResources(wantedMap, currentMap)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewArguments)
		return
	}
	adjustCmp(cmps)
	v.PlanPrinter(cmps, viewArguments)

}
