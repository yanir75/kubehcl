package configs

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert/yaml"
	"kubehcl.sh/kubehcl/internal/decode"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
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
		return "", fmt.Errorf("no module named %s", name)
	} else {
		for _, entry := range entryList {
			if entry.Version == version && len(entry.Urls) > 0 {
				return entry.Urls[0], nil
			}
		}
	}

	return "", fmt.Errorf("no matching version %s", version)
}

func replaceFirstSegment(path, newFirstSegment string) string {
	// Clean the path
	cleanPath := filepath.Clean(path)

	// Split path into parts
	parts := strings.Split(cleanPath, string(filepath.Separator))

	// Handle leading slash
	leadingSlash := strings.HasPrefix(cleanPath, string(filepath.Separator))

	// Replace first non-empty segment
	for i, part := range parts {
		if part != "" {
			parts[i] = newFirstSegment
			break
		}
	}

	// Re-join the parts
	newPath := filepath.Join(parts...)

	// Add leading slash back if it was there
	if leadingSlash {
		newPath = string(filepath.Separator) + newPath
	}

	return newPath
}

func untarFile(buff []byte, save bool, folderName string) (afero.Fs, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	appFs := afero.NewMemMapFs()

	if save {
		appFs = afero.NewOsFs()
	}

	gzf, err := gzip.NewReader(bytes.NewBuffer(buff))
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't create GZIP reader\nmake sure the file type is gz",
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
			d := replaceFirstSegment(name, folderName)
			err := appFs.Mkdir(d, 0755)
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
			d := replaceFirstSegment(filepath.Dir(name), folderName)
			name := d + string(filepath.Separator) + filepath.Base(name)
			err = appFs.MkdirAll(d, 0755)
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

func pullHttp(r *decode.DecodedRepo, name string, version string, save bool) (afero.Fs, hcl.Diagnostics) {
	httpClient, diags := NewHttpClient(r)
	if diags.HasErrors() {
		return nil, diags
	}

	res, diags := DoRequest(r, DOWNLOADINDEXFILE, httpClient, "")
	if diags.HasErrors() {
		return nil, diags
	}
	var entries Entrys
	err := yaml.Unmarshal(res.Bytes(), &entries)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Yaml is invalid",
			Detail:   fmt.Sprintf("%s \nis invalid", err.Error()),
		})
		return nil, diags
	}

	if u, err := entries.contains(name, version); err == nil {
		res, diags = DoRequest(r, "", httpClient, u)
		if diags.HasErrors() {
			return nil, diags
		}
		return untarFile(res.Bytes(), save, name)
	} else {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't download module",
			Detail:   fmt.Sprintf("Module couldn't be downloaded, error: %s", err.Error()),
		})
		return nil, diags
	}

}

func pullOci(r *decode.DecodedRepo, tag string, save bool) (afero.Fs, hcl.Diagnostics) {
	repository, err := remote.NewRepository(r.Url)
	var diags hcl.Diagnostics
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't create repository",
			Detail:   fmt.Sprintf("Repository is %s invalid, error: %s", r.Name, err.Error()),
			Subject:  &r.DeclRange,
		})
		return nil, diags
	}

	authClient, diags := NewAuthClient(r, repository.Reference.Registry)
	if diags.HasErrors() {
		return nil, diags
	}
	repository.Client = authClient
	repository.PlainHTTP = r.PlainHttp

	_, fetchedManifestContent, err := oras.FetchBytes(context.Background(), repository, tag, oras.DefaultFetchBytesOptions)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't pull module",
			Detail:   fmt.Sprintf("Tag %s cannot be pulled error: %s", tag, err.Error()),
			Subject:  &r.DeclRange,
		})
		return nil, diags
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
		return nil, diags
	}

	layerContent, err := content.FetchAll(context.Background(), repository, manifest.Layers[0])

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to fetch layer manifest",
			Detail:   fmt.Sprintf("Wan't able to retreive layer %s, error: %s", manifest.Layers[0].URLs, err.Error()),
		})
		return nil, diags
	}

	return untarFile(layerContent, save, tag)
}

func Pull(version string, repoConfigFile string, repoName string, tag string, save bool) (afero.Fs, hcl.Diagnostics) {

	repos, diags := DecodeRepos(repoConfigFile)
	if diags.HasErrors() {
		return nil, diags
	}

	repo, ok := repos[repoName]
	if !ok {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Repository doesn't exist",
			Detail:   fmt.Sprintf("%s doesn't exist please add or use other repo name.\n In order to see the repositories please use kubehcl repo list", repoName),
		})
		return nil, diags
	}

	if (repo.Protocol == "https" || repo.Protocol == "http") && version == "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Https repo must include version",
			Detail:   fmt.Sprintf("repository %s uses protocol https packages have version, please add --version", repoName),
			Subject:  &repo.DeclRange,
		})
		return nil, diags
	}

	switch repo.Protocol {
	case "oci":
		appFs, diags := pullOci(repo, tag, save)
		if diags.HasErrors() {
			return nil, diags
		}
		return appFs, diags
	case "https":
		appFs, diags := pullHttp(repo, tag, version, save)
		if diags.HasErrors() {
			return nil, diags
		}
		return appFs, diags
	case "http":
		appFs, diags := pullHttp(repo, tag, version, save)
		if diags.HasErrors() {
			return nil, diags
		}
		return appFs, diags
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Protocol is invalid",
			Detail:   fmt.Sprintf("repository %s uses protocol %s which is invalid", repoName, repo.Protocol),
		})
		return nil, diags
	}

}

type BasicAuthTransport struct {
	Transport *http.Transport
	Username  string
	Passowrd  string
}

func (b *BasicAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if b.Username != "" && b.Passowrd != "" {
		r.SetBasicAuth(b.Username, b.Passowrd)
	}
	return b.Transport.RoundTrip(r)
}

func NewHttpClient(opts *decode.DecodedRepo) (*http.Client, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	cfg := &tls.Config{
		InsecureSkipVerify: opts.InsecureSkipTLSverify,
	}
	if opts.KeyFile != "" && opts.CertFile != "" {
		cert, err := tls.LoadX509KeyPair(opts.CertFile, opts.KeyFile)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Couldn't load certificate",
				Detail:   fmt.Sprint(err.Error()),
			})
			return nil, diags
		}
		cfg.Certificates = append(cfg.Certificates, cert)
	}
	if opts.CaFile != "" {
		caCert, err := os.ReadFile(opts.CaFile)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Couldn't load CA certificate",
				Detail:   fmt.Sprint(err.Error()),
			})
			return nil, diags
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		cfg.RootCAs = caCertPool
	}

	httpClient := &http.Client{
		Transport: &BasicAuthTransport{
			Transport: &http.Transport{
				TLSClientConfig: cfg,
			},
			Username: opts.Username,
			Passowrd: opts.Password,
		},
	}

	return httpClient, diags
}

func NewAuthClient(opts *decode.DecodedRepo, regName string) (*auth.Client, hcl.Diagnostics) {
	httpClient, diags := NewHttpClient(opts)
	if diags.HasErrors() {
		return nil, diags
	}

	authClient := &auth.Client{
		Client: httpClient,
		Cache:  auth.NewCache(),
		Credential: auth.StaticCredential(regName, auth.Credential{
			Username: opts.Username,
			Password: opts.Password,
		}),
	}

	return authClient, diags
}

func DoRequest(opts *decode.DecodedRepo, path string, httpClient *http.Client, fullUrl string) (*bytes.Buffer, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var err error
	var req *http.Request

	if fullUrl == "" {
		req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s://%s/%s", opts.Protocol, opts.Url, path), nil)
	} else if !strings.Contains(fullUrl, opts.Url) {
		req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s://%s/%s", opts.Protocol, opts.Url, fullUrl), nil)
	} else {
		req, err = http.NewRequest(http.MethodGet, fullUrl, nil)
	}

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't create request",
			Detail:   fmt.Sprintf("Request couldn't be created err: %s", err.Error()),
		})
		return nil, diags
	}

	req.Header.Set("User-Agent", "kubehcl")

	if opts.Username != "" && opts.Password != "" {
		req.SetBasicAuth(opts.Username, opts.Password)
	}

	resp, err := httpClient.Do(req)

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Request failed",
			Detail:   fmt.Sprintf("Request couldn't be created err: %s", err.Error()),
		})
		return nil, diags
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Request failed",
			Detail:   fmt.Sprintf("Status code is %d", resp.StatusCode),
		})
		return nil, diags
	}

	buf := bytes.NewBuffer(nil)

	_, err = io.Copy(buf, resp.Body)

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't copy body to buff",
			Detail:   fmt.Sprint(err.Error()),
		})
		return nil, diags
	}

	return buf, diags
}
