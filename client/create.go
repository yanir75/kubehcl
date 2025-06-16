package client

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed files/*
var files embed.FS

func cacheDir(outputDir string) {

	fs.WalkDir(files, "files", func(path string, d fs.DirEntry, err error) error {
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
}
