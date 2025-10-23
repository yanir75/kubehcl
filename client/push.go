package client

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"kubehcl.sh/kubehcl/internal/configs"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
)

const (
	KUBEHCLTYPE  = "application/kubehcl.tar"
	ARTIFACTTYPE = "application/kubehcl+type"
)

func parsePushArgs(args []string) (string, string, string, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	if len(args) != 3 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Required arguments are :[folder, repo, tag]",
		})
		return "", "", "", diags
	}
	return args[0], args[1], args[2], diags

}

func CreateTar(sourceDir string) ([]byte, error) {
	buf := new(bytes.Buffer)

	gzipWriter := gzip.NewWriter(buf)
	tarWriter := tar.NewWriter(gzipWriter)

	sourceDir = filepath.Clean(sourceDir)

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		baseDir := filepath.Base(sourceDir)
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		relPath = filepath.Join(baseDir, relPath)

		if relPath == "" {
			return nil // skip root dir
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		header.Name = relPath // Preserve relative path

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func() { _ = file.Close() }()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Close writers in correct order
	if err := tarWriter.Close(); err != nil {
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func Push(defs *settings.EnvSettings, viewDef *view.ViewArgs, args []string) hcl.Diagnostics {
	folder, repoName, tag, diags := parsePushArgs(args)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return diags
	}
	annotations, diags := configs.DecodeIndexFile(fmt.Sprintf("%s%s%s", folder, string(filepath.Separator), configs.INDEXVARSFILE))
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return diags
	}

	repos, diags := configs.DecodeRepos(defs.RepositoryConfig)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return diags
	}

	decodedRepo, ok := repos[repoName]
	if !ok {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Repository doesn't exist",
			Detail:   fmt.Sprintf("%s doesn't exist please add or use other repo name.\n In order to see the repositories please use kubehcl repo list", repoName),
		})
		v.DiagPrinter(diags, viewDef)
		return diags
	}

	buff, err := CreateTar(folder)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't tar folder",
			Detail:   fmt.Sprintf("Failed to tar folder, error: %s", err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return diags
	}

	ctx := context.Background()
	repo, err := remote.NewRepository(decodedRepo.Url)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't init repository",
			Detail:   fmt.Sprintf("Failed to init repository, error: %s", err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return diags
	}
	repo.Client, diags = configs.NewAuthClient(decodedRepo, repo.Reference.Registry)
	repo.PlainHTTP = decodedRepo.PlainHttp
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return diags
	}

	layerDescriptor, err := oras.PushBytes(ctx, repo, KUBEHCLTYPE, buff)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't push data to the repository",
			Detail:   fmt.Sprintf("Failed to push data , error: %s", err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return diags
	}

	packOpts := oras.PackManifestOptions{
		Layers:              []ocispec.Descriptor{layerDescriptor},
		ManifestAnnotations: annotations,
	}

	desc, err := oras.PackManifest(ctx, repo, oras.PackManifestVersion1_1, ARTIFACTTYPE, packOpts)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't pack manifest",
			Detail:   fmt.Sprintf("Failed to pack the manifest , error: %s", err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return diags
	}

	err = repo.Tag(ctx, desc, tag)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to tag the module pushed to the OCI",
			Detail:   fmt.Sprintf("Failed to tag the module , error: %s", err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return diags
	}
	return diags

}
