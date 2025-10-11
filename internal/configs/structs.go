package configs

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/afero"
	"kubehcl.sh/kubehcl/internal/decode"
)

// Need to create a deep copy option for myself

type ModuleCall struct {
	decode.Deployable
	Source  hcl.Expression `json:"Source"`
	Version hcl.Expression `json:"Version"`
	Scope afero.Fs
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
	Scope          afero.Fs
}

type ModuleList []*Module
