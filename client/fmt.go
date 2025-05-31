package client

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)


func formatTypeExpr(tokens hclwrite.Tokens) hclwrite.Tokens {
	switch len(tokens) {
	case 1:
		kwTok := tokens[0]
		if kwTok.Type != hclsyntax.TokenIdent {
			// Not a single type keyword, then.
			return tokens
		}

		// Collection types without an explicit element type mean
		// the element type is "any", so we'll normalize that.
		switch string(kwTok.Bytes) {
		case "list", "map", "set":
			return hclwrite.Tokens{
				kwTok,
				{
					Type:  hclsyntax.TokenOParen,
					Bytes: []byte("("),
				},
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("any"),
				},
				{
					Type:  hclsyntax.TokenCParen,
					Bytes: []byte(")"),
				},
			}
		default:
			return tokens
		}

	case 3:
		// A pre-0.12 legacy quoted string type, like "string".
		oQuote := tokens[0]
		strTok := tokens[1]
		cQuote := tokens[2]
		if oQuote.Type != hclsyntax.TokenOQuote || strTok.Type != hclsyntax.TokenQuotedLit || cQuote.Type != hclsyntax.TokenCQuote {
			// Not a quoted string sequence, then.
			return tokens
		}

		// Because this quoted syntax is from Terraform 0.11 and
		// earlier, which didn't have the idea of "any" as an,
		// element type, we use string as the default element
		// type. That will avoid oddities if somehow the configuration
		// was relying on numeric values being auto-converted to
		// string, as 0.11 would do. This mimics what terraform
		// 0.12upgrade used to do, because we'd found real-world
		// modules that were depending on the auto-stringing.)
		switch string(strTok.Bytes) {
		case "string":
			return hclwrite.Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("string"),
				},
			}
		case "list":
			return hclwrite.Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("list"),
				},
				{
					Type:  hclsyntax.TokenOParen,
					Bytes: []byte("("),
				},
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("string"),
				},
				{
					Type:  hclsyntax.TokenCParen,
					Bytes: []byte(")"),
				},
			}
		case "map":
			return hclwrite.Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("map"),
				},
				{
					Type:  hclsyntax.TokenOParen,
					Bytes: []byte("("),
				},
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("string"),
				},
				{
					Type:  hclsyntax.TokenCParen,
					Bytes: []byte(")"),
				},
			}
		default:
			// Something else we're not expecting, then.
			return tokens
		}
	default:
		return tokens
	}
}

func trimNewlines(tokens hclwrite.Tokens) hclwrite.Tokens {
	if len(tokens) == 0 {
		return nil
	}
	var start, end int
	for start = 0; start < len(tokens); start++ {
		if tokens[start].Type != hclsyntax.TokenNewline {
			break
		}
	}
	for end = len(tokens); end > 0; end-- {
		if tokens[end-1].Type != hclsyntax.TokenNewline {
			break
		}
	}
	return tokens[start:end]
}

func formatValueExpr(tokens hclwrite.Tokens) hclwrite.Tokens {
	if len(tokens) < 5 {
		// Can't possibly be a "${ ... }" sequence without at least enough
		// tokens for the delimiters and one token inside them.
		return tokens
	}
	oQuote := tokens[0]
	oBrace := tokens[1]
	cBrace := tokens[len(tokens)-2]
	cQuote := tokens[len(tokens)-1]
	if oQuote.Type != hclsyntax.TokenOQuote || oBrace.Type != hclsyntax.TokenTemplateInterp || cBrace.Type != hclsyntax.TokenTemplateSeqEnd || cQuote.Type != hclsyntax.TokenCQuote {
		// Not an interpolation sequence at all, then.
		return tokens
	}

	inside := tokens[2 : len(tokens)-2]

	// We're only interested in sequences that are provable to be single
	// interpolation sequences, which we'll determine by hunting inside
	// the interior tokens for any other interpolation sequences. This is
	// likely to produce false negatives sometimes, but that's better than
	// false positives and we're mainly interested in catching the easy cases
	// here.
	quotes := 0
	for _, token := range inside {
		if token.Type == hclsyntax.TokenOQuote {
			quotes++
			continue
		}
		if token.Type == hclsyntax.TokenCQuote {
			quotes--
			continue
		}
		if quotes > 0 {
			// Interpolation sequences inside nested quotes are okay, because
			// they are part of a nested expression.
			// "${foo("${bar}")}"
			continue
		}
		if token.Type == hclsyntax.TokenTemplateInterp || token.Type == hclsyntax.TokenTemplateSeqEnd {
			// We've found another template delimiter within our interior
			// tokens, which suggests that we've found something like this:
			// "${foo}${bar}"
			// That isn't unwrappable, so we'll leave the whole expression alone.
			return tokens
		}
		if token.Type == hclsyntax.TokenQuotedLit {
			// If there's any literal characters in the outermost
			// quoted sequence then it is not unwrappable.
			return tokens
		}
	}

	// If we got down here without an early return then this looks like
	// an unwrappable sequence, but we'll trim any leading and trailing
	// newlines that might result in an invalid result if we were to
	// naively trim something like this:
	// "${
	//    foo
	// }"
	trimmed := trimNewlines(inside)

	// Finally, we check if the unwrapped expression is on multiple lines. If
	// so, we ensure that it is surrounded by parenthesis to make sure that it
	// parses correctly after unwrapping. This may be redundant in some cases,
	// but is required for at least multi-line ternary expressions.
	isMultiLine := false
	hasLeadingParen := false
	hasTrailingParen := false
	for i, token := range trimmed {
		switch {
		case i == 0 && token.Type == hclsyntax.TokenOParen:
			hasLeadingParen = true
		case token.Type == hclsyntax.TokenNewline:
			isMultiLine = true
		case i == len(trimmed)-1 && token.Type == hclsyntax.TokenCParen:
			hasTrailingParen = true
		}
	}
	if isMultiLine && (!hasLeadingParen || !hasTrailingParen) {
		wrapped := make(hclwrite.Tokens, 0, len(trimmed)+2)
		wrapped = append(wrapped, &hclwrite.Token{
			Type:  hclsyntax.TokenOParen,
			Bytes: []byte("("),
		})
		wrapped = append(wrapped, trimmed...)
		wrapped = append(wrapped, &hclwrite.Token{
			Type:  hclsyntax.TokenCParen,
			Bytes: []byte(")"),
		})

		return wrapped
	}

	return trimmed
}

func formatBody(body *hclwrite.Body, inBlocks []string) {
	attrs := body.Attributes()
	for name, attr := range attrs {
		if len(inBlocks) == 1 && inBlocks[0] == "variable" && name == "type" {
			cleanedExprTokens := formatTypeExpr(attr.Expr().BuildTokens(nil))
			body.SetAttributeRaw(name, cleanedExprTokens)
			continue
		}
		cleanedExprTokens := formatValueExpr(attr.Expr().BuildTokens(nil))
		body.SetAttributeRaw(name, cleanedExprTokens)
	}

	blocks := body.Blocks()
	for _, block := range blocks {
		// Normalize the label formatting, removing any weird stuff like
		// interleaved inline comments and using the idiomatic quoted
		// label syntax.
		block.SetLabels(block.Labels())

		inBlocks := append(inBlocks, block.Type())
		formatBody(block.Body(), inBlocks)
	}
}


var (
	fileExt = []string{
		".hcl",
		".tfvars",
	}
)

func ignoreHiddenFiles(path string) bool{
	return strings.HasPrefix(path,".")  || strings.HasPrefix(path,"~")
}

func checkFileEndings(path string) bool {
	return slices.Contains(fileExt,filepath.Ext(path))
}

func validFile(path string) bool{
	return checkFileEndings(path) && !ignoreHiddenFiles(path)
} 

func fmtDir(path string, recursive bool) hcl.Diagnostics{
	var diags hcl.Diagnostics
	entries, err := os.ReadDir(path)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: fmt.Sprintf("Directory is not valid %s",path),
			Detail: fmt.Sprintf("Couldn't read directory %s got error %s", path, err.Error()),
		})
		return diags
	}
	for _,entry := range entries {
		if entry.IsDir() {
			if recursive {
				fmtDir(filepath.Join(path,entry.Name()),recursive)
			}
		} else {
			if !validFile(entry.Name()){
				continue
			}
			name := filepath.Join(path,entry.Name())
			input, err := os.Open(name)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary: fmt.Sprintf("File is not valid %s",name),
					Detail: fmt.Sprintf("Couldn't read directory %s got error %s", name, err.Error()),
				})
			}
			defer input.Close()
					
			src, err := io.ReadAll(input)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary: fmt.Sprintf("File is not valid %s",name),
					Detail: fmt.Sprintf("Couldn't read File %s got error %s", name, err.Error()),
				})
			}
			f,parseDiags := hclwrite.ParseConfig(src,name,hcl.InitialPos)
			diags = append(diags, parseDiags...)
			if diags.HasErrors() {
				return diags
			}
			formatBody(f.Body(),nil)
			if !bytes.Equal(f.Bytes(),src) {
				overwriteF,err := os.Create(name)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary: fmt.Sprintf("Couldn't overwrite file",name),
						Detail: fmt.Sprintf("Couldn't overwrite File %s got error %s", name, err.Error()),
					})
				}

				_, err =overwriteF.Write(f.Bytes())

				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary: fmt.Sprintf("Couldn't write to file",name),
						Detail: fmt.Sprintf("Couldn't write to File %s got error %s", name, err.Error()),
					})
				}
			}



		}
	}

	return nil
}