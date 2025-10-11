package client

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert/yaml"
	"kubehcl.sh/kubehcl/internal/configs"
	"kubehcl.sh/kubehcl/internal/decode"
	"kubehcl.sh/kubehcl/internal/view"
	"kubehcl.sh/kubehcl/settings"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry/remote"
)

type Entry struct {
	Urls    []string `yaml:"urls"`
	Version string   `yaml:"version"`
}

type Entrys struct {
	EntryMap map[string][]*Entry `yaml:"entries"`
}

func (e Entrys) contains(name, version string) (string, error) {
	if entryList, ok := e.EntryMap[name]; !ok {
		return "", fmt.Errorf("No module named %s", name)
	} else {
		for _, entry := range entryList {
			if entry.Version == version && len(entry.Urls) > 0 {
				return entry.Urls[0], nil
			}
		}
	}

	return "", fmt.Errorf("No matching version %s", version)
}

func parsePullArgs(args []string) (string, string, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	if len(args) != 2 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Required arguments are :[repo, tag/name]",
		})
		return "", "", diags
	}
	return args[0], args[1], diags

}

func untarFile(buff []byte, save bool) (afero.Fs, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	appFs := afero.NewMemMapFs()

	if save {
		appFs = afero.NewOsFs()
	}

	gzf, err := gzip.NewReader(bytes.NewBuffer(buff))
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't create GZIP reader",
			Detail:   fmt.Sprintf("Gzip reader is invalid error: %s", err.Error()),
		})
		return appFs, diags
	}

	tarReader := tar.NewReader(gzf)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Couldn't get tar next",
				Detail:   fmt.Sprintf("Tar next is invalid: %s", err.Error()),
			})
			break
		}

		name := header.Name

		switch header.Typeflag {
		case tar.TypeDir: // = directory
			err := appFs.Mkdir(name, 0755)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Couldn't create folder",
					Detail:   fmt.Sprintf("Folder \"%s\" couldn't be created error: %s", name, err.Error()),
				})
				return appFs, diags
			}
		case tar.TypeReg: // = regular file
			data := make([]byte, header.Size)
			_, err := tarReader.Read(data)
			if err != nil && err != io.EOF {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Couldn't create file",
					Detail:   fmt.Sprintf("File \"%s\" couldn't be created error: %s", name, err.Error()),
				})
				return appFs, diags

			}

			err = appFs.MkdirAll(filepath.Dir(name), 0755)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Couldn't create folder",
					Detail:   fmt.Sprintf("Folder \"%s\" couldn't be created error: %s", name, err.Error()),
				})
				return appFs, diags
			}
			err = afero.WriteFile(appFs, name, data, 0755)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Couldn't write to file",
					Detail:   fmt.Sprintf("Can not write to file \"%s\" error: %s", name, err.Error()),
				})
				return appFs, diags
			}

		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to understand filetype",
				Detail:   fmt.Sprintf("Unable to understand the file type %s, filename %s", string(header.Typeflag), name),
			})
			return appFs, diags

		}
	}

	return appFs, diags
}

func pullHttp(r *decode.DecodedRepo, name string, version string, save bool) hcl.Diagnostics {
	opts := repoToOpts(r)
	httpClient, diags := newHttpClient(opts)
	if diags.HasErrors() {
		return diags
	}

	res, diags := doRequest(opts, INDEXFILE, httpClient, "")
	if diags.HasErrors() {
		return diags
	}
	var entries Entrys
	err := yaml.Unmarshal(res.Bytes(), &entries)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Yaml is invalid",
			Detail:   fmt.Sprintf("%s \nis invalid", err.Error()),
		})
		return diags
	}

	if u, err := entries.contains(name, version); err == nil {
		res, diags = doRequest(opts, "", httpClient, u)
		if diags.HasErrors() {
			return diags
		}
		_, diags = untarFile(res.Bytes(), save)
	} else {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't download module",
			Detail:   fmt.Sprintf("Module couldn't be downloaded, error: %s", err.Error()),
		})
	}

	return diags
}

func pullOci(r *decode.DecodedRepo, tag string, save bool) hcl.Diagnostics {
	repository, err := remote.NewRepository(r.Url)
	var diags hcl.Diagnostics
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't create repository",
			Detail:   fmt.Sprintf("Repository is %s invalid, error: %s", r.Name, err.Error()),
			Subject:  &r.DeclRange,
		})
		return diags
	}

	authClient, diags := newAuthClient(repoToOpts(r), repository.Reference.Registry)
	if diags.HasErrors() {
		return diags
	}
	repository.Client = authClient

	_, fetchedManifestContent, err := oras.FetchBytes(context.Background(), repository, tag, oras.DefaultFetchBytesOptions)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't pull module",
			Detail:   fmt.Sprintf("Tag %s cannot be pulled error: %s", tag, err.Error()),
			Subject:  &r.DeclRange,
		})
		return diags
	}
	var manifest ocispec.Manifest
	if err := json.Unmarshal(fetchedManifestContent, &manifest); err != nil {
		panic(err)
	}

	if len(manifest.Layers) != 1 {
		diags = append(diags,
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Too many layers to the manifest expecting only 1",
				Detail:   fmt.Sprintf("Manifest has %d layers", len(manifest.Layers)),
			})
		return diags
	}

	layerContent, err := content.FetchAll(context.Background(), repository, manifest.Layers[0])

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to fetch layer manifest",
			Detail:   fmt.Sprintf("Wan't able to retreive layer %s, error: %s", manifest.Layers[0].URLs, err.Error()),
		})
		return diags
	}

	_, diags = untarFile(layerContent, save)

	return diags
}

func Pull(version string, envSettings *settings.EnvSettings, viewDef *view.ViewArgs, args []string) {
	repoName, tag, diags := parsePullArgs(args)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return
	}

	repos, diags := configs.DecodeRepos(envSettings.RepositoryConfig)
	if diags.HasErrors() {
		v.DiagPrinter(diags, viewDef)
		return
	}

	repo, ok := repos[repoName]
	if !ok {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Repository doesn't exist",
			Detail:   fmt.Sprintf("%s doesn't exist please add or use other repo name.\n In order to see the repositories please use kubehcl repo list", repoName),
		})
		v.DiagPrinter(diags, viewDef)
		return
	}

	if (repo.Protocol == "https" || repo.Protocol == "http") && version == "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Https repo must include version",
			Detail:   fmt.Sprintf("repository %s uses protocol https packages have version, please add --version", repoName),
			Subject:  &repo.DeclRange,
		})
		v.DiagPrinter(diags, viewDef)
		return
	}

	switch repo.Protocol {
	case "oci":
		diags = pullOci(repo, tag, true)
		if diags.HasErrors() {
			v.DiagPrinter(diags, viewDef)
			return
		}
	case "https":
		diags = pullHttp(repo, tag, version, true)
		if diags.HasErrors() {
			v.DiagPrinter(diags, viewDef)
			return
		}
	case "http":
		diags = pullHttp(repo, tag, version, true)
		if diags.HasErrors() {
			v.DiagPrinter(diags, viewDef)
			return
		}

	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Protocol is invalid",
			Detail:   fmt.Sprintf("repository %s uses protocol %s which is invalid", repoName, repo.Protocol),
		})
		v.DiagPrinter(diags, viewDef)
		return
	}

}
