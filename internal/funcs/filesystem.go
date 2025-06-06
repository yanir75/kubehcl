/*
// SPDX-License-Identifier: MPL-2.0
This file was copied from https://github.com/opentofu/opentofu and retains its' original license: https://www.mozilla.org/en-US/MPL/2.0/
*/
// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package funcs

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// MakeFileFunc constructs a function that takes a file path and returns the
// contents of that file, either directly as a string (where valid UTF-8 is
// required) or as a string containing base64 bytes.
func MakeFileFunc(baseDir string, encBase64 bool) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:        "path",
				Type:        cty.String,
				AllowMarked: true,
			},
		},
		Type:         function.StaticReturnType(cty.String),
		RefineResult: refineNotNull,
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			pathArg, pathMarks := args[0].Unmark()
			path := pathArg.AsString()
			src, err := readFileBytes(baseDir, path, pathMarks)
			if err != nil {
				err = function.NewArgError(0, err)
				return cty.UnknownVal(cty.String), err
			}

			switch {
			case encBase64:
				enc := base64.StdEncoding.EncodeToString(src)
				return cty.StringVal(enc).WithMarks(pathMarks), nil
			default:
				if !utf8.Valid(src) {
					return cty.UnknownVal(cty.String), fmt.Errorf("contents of %s are not valid UTF-8; use the filebase64 function to obtain the Base64 encoded contents or the other file functions (e.g. filemd5, filesha256) to obtain file hashing results instead", redactIfSensitive(path, pathMarks))
				}
				return cty.StringVal(string(src)).WithMarks(pathMarks), nil
			}
		},
	})
}

func templateMaxRecursionDepth() (int, error) {
	envkey := "TF_TEMPLATE_RECURSION_DEPTH"
	val := os.Getenv(envkey)
	if val != "" {
		i, err := strconv.Atoi(val)
		if err != nil {
			return -1, fmt.Errorf("invalid value for %s: %w", envkey, err)
		}
		return i, nil
	}
	return 1024, nil // Sane Default
}

type ErrorTemplateRecursionLimit struct {
	sources []string
}

func (err ErrorTemplateRecursionLimit) Error() string {
	trace := make([]string, 0)
	maxTrace := 16

	// Look for repetition in the first N sources
	for _, source := range err.sources[:min(maxTrace, len(err.sources))] {
		looped := false
		for _, st := range trace {
			if st == source {
				// Repeated source, probably a loop.  TF_LOG=debug will contain the full trace.
				looped = true
				break
			}
		}

		trace = append(trace, source)

		if looped {
			break
		}
	}

	log.Printf("[DEBUG] Template Stack (%d): %s", len(err.sources)-1, err.sources[len(err.sources)-1])

	return fmt.Sprintf("maximum recursion depth %d reached in %s ... ", len(err.sources)-1, strings.Join(trace, ", "))
}

// MakeTemplateFileFunc constructs a function that takes a file path and
// an arbitrary object of named values and attempts to render the referenced
// file as a template using HCL template syntax.
//
// The template itself may recursively call other functions so a callback
// must be provided to get access to those functions. The template cannot,
// however, access any variables defined in the scope: it is restricted only to
// those variables provided in the second function argument, to ensure that all
// dependencies on other graph nodes can be seen before executing this function.
//
// As a special exception, a referenced template file may call the templatefile
// function, with a recursion depth limit providing an error when reached
func MakeTemplateFileFunc(baseDir string, funcsCb func() map[string]function.Function) function.Function {
	return makeTemplateFileFuncImpl(baseDir, funcsCb, 0)
}
func makeTemplateFileFuncImpl(baseDir string, funcsCb func() map[string]function.Function, depth int) function.Function {
	params := []function.Parameter{
		{
			Name:        "path",
			Type:        cty.String,
			AllowMarked: true,
		},
		{
			Name: "vars",
			Type: cty.DynamicPseudoType,
		},
	}

	loadTmpl := func(path string, marks cty.ValueMarks) (hcl.Expression, error) {
		maxDepth, err := templateMaxRecursionDepth()
		if err != nil {
			return nil, err
		}
		if depth > maxDepth {
			// Sources will unwind up the stack
			return nil, ErrorTemplateRecursionLimit{}
		}

		// We re-use File here to ensure the same filename interpretation
		// as it does, along with its other safety checks.
		templateValue, err := File(baseDir, cty.StringVal(path).WithMarks(marks))
		if err != nil {
			return nil, err
		}

		// unmark the template ready to be handled
		templateValue, _ = templateValue.Unmark()

		expr, diags := hclsyntax.ParseTemplate([]byte(templateValue.AsString()), path, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return nil, diags
		}

		return expr, nil
	}

	funcsCbDepth := func() map[string]function.Function {
		givenFuncs := funcsCb() // this callback indirection is to avoid chicken/egg problems
		funcs := make(map[string]function.Function, len(givenFuncs))
		for name, fn := range givenFuncs {
			if name == "templatefile" {
				// Increment the recursion depth counter
				funcs[name] = makeTemplateFileFuncImpl(baseDir, funcsCb, depth+1)
				continue
			}
			funcs[name] = fn
		}
		return funcs
	}

	return function.New(&function.Spec{
		Params: params,
		Type: func(args []cty.Value) (cty.Type, error) {
			if !(args[0].IsKnown() && args[1].IsKnown()) {
				return cty.DynamicPseudoType, nil
			}

			// We'll render our template now to see what result type it produces.
			// A template consisting only of a single interpolation an potentially
			// return any type.

			pathArg, pathMarks := args[0].Unmark()
			expr, err := loadTmpl(pathArg.AsString(), pathMarks)
			if err != nil {
				return cty.DynamicPseudoType, err
			}

			// This is safe even if args[1] contains unknowns because the HCL
			// template renderer itself knows how to short-circuit those.
			val, err := renderTemplate(expr, args[1], funcsCbDepth())
			return val.Type(), err
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			pathArg, pathMarks := args[0].Unmark()
			expr, err := loadTmpl(pathArg.AsString(), pathMarks)
			if err != nil {
				return cty.DynamicVal, err
			}

			result, err := renderTemplate(expr, args[1], funcsCbDepth())
			return result.WithMarks(pathMarks), err
		},
	})
}

// MakeFileExistsFunc constructs a function that takes a path
// and determines whether a file exists at that path
func MakeFileExistsFunc(baseDir string) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:        "path",
				Type:        cty.String,
				AllowMarked: true,
			},
		},
		Type:         function.StaticReturnType(cty.Bool),
		RefineResult: refineNotNull,
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			pathArg, pathMarks := args[0].Unmark()
			path := pathArg.AsString()
			path, err := homedir.Expand(path)
			if err != nil {
				return cty.UnknownVal(cty.Bool), fmt.Errorf("failed to expand ~: %w", err)
			}

			if !filepath.IsAbs(path) {
				path = filepath.Join(baseDir, path)
			}

			// Ensure that the path is canonical for the host OS
			path = filepath.Clean(path)

			fi, err := os.Stat(path)
			if err != nil {
				if os.IsNotExist(err) {
					return cty.False.WithMarks(pathMarks), nil
				}
				return cty.UnknownVal(cty.Bool), fmt.Errorf("failed to stat %s", redactIfSensitive(path, pathMarks))
			}

			if fi.Mode().IsRegular() {
				return cty.True.WithMarks(pathMarks), nil
			}

			// The Go stat API only provides convenient access to whether it's
			// a directory or not, so we need to do some bit fiddling to
			// recognize other irregular file types.
			filename := redactIfSensitive(path, pathMarks)
			fileType := fi.Mode().Type()
			switch {
			case (fileType & os.ModeDir) != 0:
				err = function.NewArgErrorf(1, "%s is a directory, not a file", filename)
			case (fileType & os.ModeDevice) != 0:
				err = function.NewArgErrorf(1, "%s is a device node, not a regular file", filename)
			case (fileType & os.ModeNamedPipe) != 0:
				err = function.NewArgErrorf(1, "%s is a named pipe, not a regular file", filename)
			case (fileType & os.ModeSocket) != 0:
				err = function.NewArgErrorf(1, "%s is a unix domain socket, not a regular file", filename)
			default:
				// If it's not a type we recognize then we'll just return a
				// generic error message. This should be very rare.
				err = function.NewArgErrorf(1, "%s is not a regular file", filename)

				// Note: os.ModeSymlink should be impossible because we used
				// os.Stat above, not os.Lstat.
			}

			return cty.False, err
		},
	})
}

// MakeFileSetFunc constructs a function that takes a glob pattern
// and enumerates a file set from that pattern
func MakeFileSetFunc(baseDir string) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:        "path",
				Type:        cty.String,
				AllowMarked: true,
			},
			{
				Name:        "pattern",
				Type:        cty.String,
				AllowMarked: true,
			},
		},
		Type:         function.StaticReturnType(cty.Set(cty.String)),
		RefineResult: refineNotNull,
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			pathArg, pathMarks := args[0].Unmark()
			path := pathArg.AsString()
			patternArg, patternMarks := args[1].Unmark()
			pattern := patternArg.AsString()

			marks := []cty.ValueMarks{pathMarks, patternMarks}

			if !filepath.IsAbs(path) {
				path = filepath.Join(baseDir, path)
			}

			// Join the path to the glob pattern, while ensuring the full
			// pattern is canonical for the host OS. The joined path is
			// automatically cleaned during this operation.
			pattern = filepath.Join(path, pattern)

			matches, err := doublestar.FilepathGlob(pattern)
			if err != nil {
				return cty.UnknownVal(cty.Set(cty.String)), fmt.Errorf("failed to glob pattern %s: %w", redactIfSensitive(pattern, marks...), err)
			}

			var matchVals []cty.Value
			for _, match := range matches {
				fi, err := os.Stat(match)

				if err != nil {
					return cty.UnknownVal(cty.Set(cty.String)), fmt.Errorf("failed to stat %s: %w", redactIfSensitive(match, marks...), err)
				}

				if !fi.Mode().IsRegular() {
					continue
				}

				// Remove the path and file separator from matches.
				match, err = filepath.Rel(path, match)

				if err != nil {
					return cty.UnknownVal(cty.Set(cty.String)), fmt.Errorf("failed to trim path of match %s: %w", redactIfSensitive(match, marks...), err)
				}

				// Replace any remaining file separators with forward slash (/)
				// separators for cross-system compatibility.
				match = filepath.ToSlash(match)

				matchVals = append(matchVals, cty.StringVal(match))
			}

			if len(matchVals) == 0 {
				return cty.SetValEmpty(cty.String).WithMarks(marks...), nil
			}

			return cty.SetVal(matchVals).WithMarks(marks...), nil
		},
	})
}

// BasenameFunc constructs a function that takes a string containing a filesystem path
// and removes all except the last portion from it.
var BasenameFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "path",
			Type: cty.String,
		},
	},
	Type:         function.StaticReturnType(cty.String),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		return cty.StringVal(filepath.Base(args[0].AsString())), nil
	},
})

// DirnameFunc constructs a function that takes a string containing a filesystem path
// and removes the last portion from it.
var DirnameFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "path",
			Type: cty.String,
		},
	},
	Type:         function.StaticReturnType(cty.String),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		return cty.StringVal(filepath.Dir(args[0].AsString())), nil
	},
})

// AbsPathFunc constructs a function that converts a filesystem path to an absolute path
var AbsPathFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "path",
			Type: cty.String,
		},
	},
	Type:         function.StaticReturnType(cty.String),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		absPath, err := filepath.Abs(args[0].AsString())
		return cty.StringVal(filepath.ToSlash(absPath)), err
	},
})

// PathExpandFunc constructs a function that expands a leading ~ character to the current user's home directory.
var PathExpandFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "path",
			Type: cty.String,
		},
	},
	Type:         function.StaticReturnType(cty.String),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {

		homePath, err := homedir.Expand(args[0].AsString())
		return cty.StringVal(homePath), err
	},
})

func openFile(baseDir, path string) (*os.File, error) {
	path, err := homedir.Expand(path)
	if err != nil {
		return nil, fmt.Errorf("failed to expand ~: %w", err)
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
	}

	// Ensure that the path is canonical for the host OS
	path = filepath.Clean(path)

	return os.Open(path)
}

func readFileBytes(baseDir, path string, marks cty.ValueMarks) ([]byte, error) {
	f, err := openFile(baseDir, path)
	if err != nil {
		if os.IsNotExist(err) {
			// An extra OpenTofu-specific hint for this situation
			return nil, fmt.Errorf("no file exists at %s; this function works only with files that are distributed as part of the configuration source code, so if this file will be created by a resource in this configuration you must instead obtain this result from an attribute of that resource", redactIfSensitive(path, marks))
		}
		return nil, err
	}
	defer f.Close()

	src, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return src, nil
}

// File reads the contents of the file at the given path.
//
// The file must contain valid UTF-8 bytes, or this function will return an error.
//
// The underlying function implementation works relative to a particular base
// directory, so this wrapper takes a base directory string and uses it to
// construct the underlying function before calling it.
func File(baseDir string, path cty.Value) (cty.Value, error) {
	fn := MakeFileFunc(baseDir, false)
	return fn.Call([]cty.Value{path})
}

// FileExists determines whether a file exists at the given path.
//
// The underlying function implementation works relative to a particular base
// directory, so this wrapper takes a base directory string and uses it to
// construct the underlying function before calling it.
func FileExists(baseDir string, path cty.Value) (cty.Value, error) {
	fn := MakeFileExistsFunc(baseDir)
	return fn.Call([]cty.Value{path})
}

// FileSet enumerates a set of files given a glob pattern
//
// The underlying function implementation works relative to a particular base
// directory, so this wrapper takes a base directory string and uses it to
// construct the underlying function before calling it.
func FileSet(baseDir string, path, pattern cty.Value) (cty.Value, error) {
	fn := MakeFileSetFunc(baseDir)
	return fn.Call([]cty.Value{path, pattern})
}

// FileBase64 reads the contents of the file at the given path.
//
// The bytes from the file are encoded as base64 before returning.
//
// The underlying function implementation works relative to a particular base
// directory, so this wrapper takes a base directory string and uses it to
// construct the underlying function before calling it.
func FileBase64(baseDir string, path cty.Value) (cty.Value, error) {
	fn := MakeFileFunc(baseDir, true)
	return fn.Call([]cty.Value{path})
}

// Basename takes a string containing a filesystem path and removes all except the last portion from it.
//
// The underlying function implementation works only with the path string and does not access the filesystem itself.
// It is therefore unable to take into account filesystem features such as symlinks.
//
// If the path is empty then the result is ".", representing the current working directory.
func Basename(path cty.Value) (cty.Value, error) {
	return BasenameFunc.Call([]cty.Value{path})
}

// Dirname takes a string containing a filesystem path and removes the last portion from it.
//
// The underlying function implementation works only with the path string and does not access the filesystem itself.
// It is therefore unable to take into account filesystem features such as symlinks.
//
// If the path is empty then the result is ".", representing the current working directory.
func Dirname(path cty.Value) (cty.Value, error) {
	return DirnameFunc.Call([]cty.Value{path})
}

// Pathexpand takes a string that might begin with a `~` segment, and if so it replaces that segment with
// the current user's home directory path.
//
// The underlying function implementation works only with the path string and does not access the filesystem itself.
// It is therefore unable to take into account filesystem features such as symlinks.
//
// If the leading segment in the path is not `~` then the given path is returned unmodified.
func Pathexpand(path cty.Value) (cty.Value, error) {
	return PathExpandFunc.Call([]cty.Value{path})
}
