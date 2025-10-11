package configs

import (
	"github.com/hashicorp/hcl/v2"
	"kubehcl.sh/kubehcl/internal/decode"
)

// Need to create a deep copy option for myself

type ModuleCall struct {
	decode.Deployable
	Source  hcl.Expression `json:"Source"`
	Version hcl.Expression `json:"Version"`
}

type Module struct {
	Name           string          `json:"Name"`
	BackendStorage *BackendStorage `json:"BackendStorage"`
	Inputs         VariableMap     `json:"Inputs"`
	Locals         Locals          `json:"Locals"`
	Annotations    Annotations     `json:"Annotations"`
	Resources      ResourceList    `json:"Resources"`
	ModuleCalls    ModuleCallList  `json:"ModuleCalls"`
	DependsOn      []hcl.Traversal `json:"DependsOn"`
	Source         string          `json:"Source"`
	Version        string          `json:"Version"`
}

type ModuleList []*Module
