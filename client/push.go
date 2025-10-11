package client

import (
	"archive/tar"
	"bytes"
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

func CreateTar(folder string) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer func() { _ = tw.Close() }()

	err := filepath.Walk(folder, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the relative path to the source directory
		relPath, err := filepath.Rel(folder, file)
		if err != nil {
			return err
		}

		// Create a tar header
		header, err := tar.FileInfoHeader(fi, relPath)
		if err != nil {
			return err
		}

		// Ensure the name in the header is the relative path
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// If it's a regular file, copy its content
		if !fi.IsDir() {
			f, err := os.Open(file)
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()

			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func Push(defs *settings.EnvSettings, viewDef *view.ViewArgs, args []string) {
	folder, repoName, tag, diags := parsePushArgs(args)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return
	}
	annotations, diags := configs.DecodeIndexFile(fmt.Sprintf("%s%s%s", folder, string(filepath.Separator), configs.INDEXVARSFILE))
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return
	}

	repos, diags := configs.DecodeRepos(defs.RepositoryConfig)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return
	}

	decodedRepo, ok := repos[repoName]
	if !ok {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Repository doesn't exist",
			Detail:   fmt.Sprintf("%s doesn't exist please add or use other repo name.\n In order to see the repositories please use kubehcl repo list", repoName),
		})
		v.DiagPrinter(diags, viewDef)
		return
	}

	buff, err := CreateTar(folder)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't tar folder",
			Detail:   fmt.Sprintf("Failed to tar folder, error: %s", err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return
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
		return
	}
	repo.Client, diags = configs.NewAuthClient(decodedRepo, repo.Reference.Registry)

	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return
	}

	layerDescriptor, err := oras.PushBytes(ctx, repo, KUBEHCLTYPE, buff)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't push data to the repository",
			Detail:   fmt.Sprintf("Failed to push data , error: %s", err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return
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
		return
	}

	err = repo.Tag(ctx, desc, tag)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to tag the module pushed to the OCI",
			Detail:   fmt.Sprintf("Failed to tag the module , error: %s", err.Error()),
		})
		v.DiagPrinter(diags, viewDef)
		return
	}
}
