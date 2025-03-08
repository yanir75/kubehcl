package main

// import (
// 	"fmt"
// 	"os"
// 	"path/filepath"

// 	"github.com/hashicorp/hcl/v2"
// 	"github.com/zclconf/go-cty/cty"
// )

// type Module struct {
// 	Source hcl.Expression
// 	Inputs VariableList
// 	Locals Locals
// 	Annotations Annotations
// 	Resources ResourceList
// 	Module ModuleCallList
// 	SourceRange hcl.Range
// }

// type ModuleMap map[string]Module


// func (m *Module) decodeModuleFolder(ctx *hcl.EvalContext) hcl.Diagnostics{
// 	var diags hcl.Diagnostics
// 	val,valDdiags := m.Source.Value(ctx)
// 	diags = append(diags, valDdiags...)
// 	if val.Type() != cty.String {
// 		diags = append(diags,&hcl.Diagnostic{
// 			Severity:    hcl.DiagError,
// 			Summary:     "Source must be string",
// 			Detail:      fmt.Sprintf("Required string and you entered type %s",val.Type().FriendlyName()),
// 			Subject:     m.Source.Range().Ptr(),
// 			Expression:  m.Source,
// 			EvalContext: ctx,
// 		})
// 	} else{

// 	files, err := os.ReadDir(val.AsString())
// 		if err != nil {
// 			diags = append(diags,&hcl.Diagnostic{
// 				Severity:    hcl.DiagError,
// 				Summary:     "Invalid Source",
// 				Detail:      fmt.Sprintf("Can not open the given folder %s",val.AsString()),
// 				Subject:     m.Source.Range().Ptr(),
// 				Expression:  m.Source,
// 				EvalContext: ctx,
// 			})
// 		}
// 		for _,f := range files{
// 			if filepath.Ext(f.Name()) == ext{
// 				module,decodeDiags := decode(f.Name())
// 				diags = append(diags, decodeDiags...)

// 			}
// 		}
// 	}
// 	return diags
// }

