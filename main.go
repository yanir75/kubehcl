package main

import (

	// "net/http"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"helm.sh/helm/v4/pkg/kube"

	// "helm.sh/helm/v4/pkg/kube"
	// "k8s.io/cli-runtime/pkg/genericclioptions"
	// "k8s.io/client-go/rest"
	"kubehcl.sh/kubehcl/client"
	"kubehcl.sh/kubehcl/internal/configs"
	"kubehcl.sh/kubehcl/internal/dag"
	"kubehcl.sh/kubehcl/internal/decode"
	"kubehcl.sh/kubehcl/internal/view"
)

// var tasks chan Task
// var workerCtx context.Context
// var l sync.Mutex

// type Task struct {
// 	function func()
// }

// func init(){
// 	tasks = make(chan Task)
// 	workerCtx = context.Background()

// 	for i:=0; i<maxGoRountines;i++{
// 		go StartWorker(workerCtx,tasks)
// 	}
// }

// func StartWorker(ctx context.Context ,tasks <-chan Task) {
//     for {
//         select {
//             case task := <-tasks:
// 				func () {
// 					task.function()
// 				}()
// 			case <- ctx.Done():
// 				return
// 			}
//     }
// }


func main() {

	d, diags := configs.DecodeFolderAndModules(".", "root", 0)
	g := &configs.Graph{
		DecodedModule: d,
	}
	diags = append(diags, g.Init()...)
	cfg := client.New()
	var results *kube.Result = &kube.Result{}
	var mutex sync.Mutex
	validateFunc := func(v dag.Vertex) hcl.Diagnostics{
		switch tt:=v.(type){
		case *decode.DecodedResource:
			return cfg.Validate(tt)
		}
		return nil
	}
	createFunc := func(v dag.Vertex) hcl.Diagnostics{
		switch tt:=v.(type){
		case *decode.DecodedResource:
			res,createDiags := cfg.Create(tt)
			if !createDiags.HasErrors(){
				mutex.Lock()
				defer mutex.Unlock()
				results.Created = append(results.Created, res.Created...)
				results.Updated = append(results.Updated, res.Updated...)
				results.Deleted = append(results.Deleted, res.Deleted...)
			}
			return createDiags
			// fmt.Printf("%s",asdf.Created[0])
		}
		return nil
	}
	
	diags = append(diags,g.Walk(validateFunc)...)
	if !diags.HasErrors(){
		diags = append(diags,g.Walk(createFunc)...)
		cfg.DeleteResources()
		if !diags.HasErrors(){
			diags = append(diags,cfg.UpdateSecret()...)
		}
	}

	view.DiagPrinter(diags)

	// fmt.Printf("%s",g.AcyclicGraph.TopologicalOrder())
	// workerCtx.Done()
	// a := dag.Cycles()

	// for _,variable := range v {
	// 	fmt.Printf("%s",variable.Description)
	// 	fmt.Println()
	// 	fmt.Printf("%s",variable.Name)
	// 	fmt.Println()
	// 	fmt.Printf("%s",variable.Type.FriendlyName())
	// 	fmt.Println()
	// 	fmt.Printf("%s",variable.Default.Type().FriendlyName())
	// 	fmt.Println()

	// }

	// fmt.Printf("%s",i)
	// c := &Config{}
	// gohcl.DecodeBody(srcHCL.Body, nil, c)
	// for a, _ := range c.Variables {
	// 	// fmt.Println(v.Type.Expr.Variables()[0].RootName())
	// 	// fmt.Printf("%s",v.Type.Expr)
	// 	fmt.Println(a)
	// 	// k,_ := v.Default.Expr.Value(nil)
	// 	// l,_ := convert.Convert(k,cty.String)
	// 	// v.
	// }

}
