package configs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"kubehcl.sh/kubehcl/internal/decode"
)

var inputRepoBlocks = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "repo",
			LabelNames: []string{"Name"},
		},
	},
}

var inputRepoBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "Name", Required: true},
		{Name: "Url", Required: true},
		{Name: "Protocol", Required: true},
		{Name: "Username", Required: false},
		{Name: "Password", Required: false},
		{Name: "Timeout", Required: false},
		{Name: "CertFile", Required: false},
		{Name: "KeyFile", Required: false},
		{Name: "CaFile", Required: false},
		{Name: "InsecureSkipTLSverify", Required: false},
		{Name: "PlainHttp", Required: false},
		{Name: "RepoFile", Required: false},
		{Name: "RepoCache", Required: false},
	},
}

type Repo struct {
	Name                  string    // `json:"Name"`
	DeclRange             hcl.Range // `json:"DeclRange"`
	Url                   string
	Protocol              string
	Username              string
	Password              string
	Timeout               int64
	CertFile              string
	KeyFile               string
	CaFile                string
	InsecureSkipTLSverify bool
	PlainHttp             bool
	RepoFile              string
	RepoCache             string
}

type RepoMap map[string]*Repo

func getValidProtocols() []string {
	return []string{"http", "https", "oci"}
}

func (r *Repo) decode() (*decode.DecodedRepo, hcl.Diagnostics) {
	if !slices.Contains(getValidProtocols(), r.Protocol) {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Protocol is invalid",
				Detail:   fmt.Sprintf("Protocol %s is invalid please use http, https or oci", r.Protocol),
				Subject:  &r.DeclRange,
			},
		}
	}

	return &decode.DecodedRepo{
		Name:                  r.Name,
		DeclRange:             r.DeclRange,
		Url:                   r.Url,
		Protocol:              r.Protocol,
		Username:              r.Username,
		Password:              r.Password,
		Timeout:               r.Timeout,
		CertFile:              r.CertFile,
		KeyFile:               r.KeyFile,
		CaFile:                r.CaFile,
		InsecureSkipTLSverify: r.InsecureSkipTLSverify,
		PlainHttp:             r.PlainHttp,
		RepoFile:              r.RepoFile,
		RepoCache:             r.RepoCache,
	}, hcl.Diagnostics{}
}

func (r RepoMap) Decode() (decode.DecodedRepoMap, hcl.Diagnostics) {
	var decodedRepoMap decode.DecodedRepoMap = make(map[string]*decode.DecodedRepo)
	var diags hcl.Diagnostics
	for key, value := range r {
		decodedRepo, decodeDiags := value.decode()
		diags = append(diags, decodeDiags...)
		decodedRepoMap[key] = decodedRepo
	}
	return decodedRepoMap, diags
}

func DecodeRepoBlocks(blocks hcl.Blocks) (RepoMap, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var repoMap RepoMap = make(map[string]*Repo)
	for _, block := range blocks {
		repo, varDiags := decodeRepoBlock(block)
		diags = append(diags, varDiags...)
		if repo != nil {
			if _, exists := repoMap[repo.Name]; exists {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Repos must have different names",
					Detail:   fmt.Sprintf("Two repositories have the same name: %s", repo.Name),
					Subject:  &block.DefRange,
					// Context: names[variable.Name],
				})

			} else {
				repoMap[repo.Name] = repo
			}
		}
	}
	return repoMap, diags
}

func decodeRepoBlock(block *hcl.Block) (*Repo, hcl.Diagnostics) {
	var repo = &Repo{
		Name:                  block.Labels[0],
		DeclRange:             block.DefRange,
		Username:              "",
		Password:              "",
		Timeout:               120,
		CaFile:                "",
		CertFile:              "",
		KeyFile:               "",
		InsecureSkipTLSverify: false,
		PlainHttp:             false,
		RepoFile:              "",
		RepoCache:             "",
		Protocol:              "",
	}

	content, diags := block.Body.Content(inputRepoBlockSchema)
	if diags.HasErrors(){
		return &Repo{},diags
	}
	if attr, exists := content.Attributes["Url"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &repo.Url)
		diags = append(diags, valDiags...)
	} else {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: "URL must be included in repo block",
			Detail: fmt.Sprintf("URL is not included in repo block %s",block.Labels[0]),
			Subject: &block.DefRange,
		})

	}
	if attr, exists := content.Attributes["Username"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &repo.Username)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["Password"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &repo.Password)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["Timeout"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &repo.Timeout)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["CertFile"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &repo.CertFile)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["KeyFile"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &repo.KeyFile)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["CaFile"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &repo.CaFile)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["InsecureSkipTLSverify"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &repo.InsecureSkipTLSverify)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["PlainHttp"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &repo.PlainHttp)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["RepoFile"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &repo.RepoFile)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["RepoCache"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &repo.RepoCache)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["Protocol"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &repo.Protocol)
		diags = append(diags, valDiags...)
	}

	return repo, diags
}

func DecodeRepos(fileName string) (decode.DecodedRepoMap, hcl.Diagnostics) {
	if err := os.MkdirAll(filepath.Dir(fileName), 0755); err != nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Couldn't open file",
				Detail:   fmt.Sprintf("File %s couldn't be opened, %s", fileName, err.Error()),
			},
		}
	}

	input, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {

		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Couldn't open file",
				Detail:   fmt.Sprintf("File %s couldn't be opened, %s", fileName, err.Error()),
			},
		}
	}

	defer func() {
		err = input.Close()
		if err != nil {
			panic("Couldn't close the file")
		}
	}()
	_, err = input.Seek(0, 0)
	if err != nil {

		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Couldn't seek in file",
				Detail:   fmt.Sprintf("File %s couldn't be seeked, %s", fileName, err.Error()),
			},
		}
	}

	var diags hcl.Diagnostics

	src, err := io.ReadAll(input)
	if err != nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Couldn't read file",
				Detail:   fmt.Sprintf("File %s couldn't be read, %s", fileName, err.Error()),
			},
		}
	}

	srcHCL, diagsParse := parser.ParseHCL(src, fileName)
	diags = append(diags, diagsParse...)
	if diags.HasErrors() {
		return nil, diags
	}
	b, blockDiags := srcHCL.Body.Content(inputRepoBlocks)
	diags = append(diags, blockDiags...)
	if diags.HasErrors() {
		return nil, diags
	}
	repos := b.Blocks.OfType("repo")
	repoMap, diags := DecodeRepoBlocks(repos)
	if diags.HasErrors() {
		return nil, diags
	}
	m, diags := repoMap.Decode()
	return m, diags
}
