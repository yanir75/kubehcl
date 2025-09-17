package client

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

