package client

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"kubehcl.sh/kubehcl/internal/view"
)

//go:embed files/*
var files embed.FS

func cacheDir(outputDir string) {

	err := fs.WalkDir(files, "files", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		p, _ := filepath.Rel("files", path)
		target := filepath.Join(outputDir, p)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		data, err := files.ReadFile(path)
		if err != nil {
			return err
		}

		err = os.MkdirAll(filepath.Dir(target), 0755)
		if err != nil {
			return err
		}
		err = os.WriteFile(target, data, 0555)

		return err
	})
	if err != nil {
		panic("Should not get here")
	}
}

func parseCreateArgs(args []string) (string, hcl.Diagnostics) {
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

// creates a basic module
func Create(args []string, viewArguments *view.ViewArgs) {
	name, diags := parseCreateArgs(args)
	if diags.HasErrors() {
		view.DiagPrinter(diags, viewArguments)
		return
	}
	cacheDir(name)
}
