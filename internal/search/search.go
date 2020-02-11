// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The code for module matching has been adapted from
// src/cmd/go/internal/search/search.go in the Go source distribution.
// Copyright 2017 The Go Authors. All rights reserved.

// Package search implements support for searching local modules.
//
// A local module is a module whose module path, as defined in the module
// directive in go.mod, is inside $GOPATH.
package search

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/mod/modfile"
)

// Module represents a local module.
type Module struct {
	Path      string // module path
	Version   string // module version
	Main      bool   // is this the main module?
	Dir       string // directory holding files for this module
	GoMod     string // path to go.mod file for this module
	GoVersion string // go version used in module
}

// A Match represents the result of matching a single module pattern.
type Match struct {
	Pattern string    // the pattern itself
	Literal bool      // whether it is a literal (no wildcards)
	Modules []*Module // matching modules
}

// MatchModules returns all the modules that can be found under the $GOPATH
// directories matching pattern.  The pattern is a path including "...".
func MatchModules(pattern string) *Match {
	m := &Match{
		Pattern: pattern,
		Literal: isLiteral(pattern),
	}
	match := matchPattern(pattern)

	for _, dirpath := range filepath.SplitList(gopath()) {
		src := filepath.Join(dirpath, "src")
		src = filepath.Clean(src) + string(filepath.Separator)
		root := src

		filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
			if err != nil || path == src {
				return nil
			}

			if fi.IsDir() {
				return nil
			}
			name := filepath.Base(path)
			if name != "go.mod" {
				return nil
			}

			dir := filepath.Dir(path)
			mod, err := load(root, dir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "can't load module: %v\n", err)

				return nil
			}

			if !match(mod.Path) {
				return nil
			}
			m.Modules = append(m.Modules, mod)

			return nil
		})
	}

	return m
}

// Resolve resolves pattern to a local Go module.
//
// pattern can be not be an absolute or relative path.  A module path is
// resolved relative to $GOPATH.
func Resolve(pattern string) (*Module, error) {
	match := ModulePath(pattern)
	if len(match.Modules) != 1 {
		return nil, fmt.Errorf("resolve %q: unable to resolve", pattern)
	}

	return match.Modules[0], nil
}

// ModulePath returns the matching paths to use for the given command line.
func ModulePath(pattern string) *Match {
	// There is no much to to, since we don't support absolute and relative
	// filesystem paths and have only one pattern.
	if filepath.IsAbs(pattern) {
		fmt.Fprintf(os.Stderr, "not supported: %q: absolute path\n", pattern)

		return new(Match) // for consistency
	}
	if isRelativePath(pattern) {
		fmt.Fprintf(os.Stderr, "not supported: %q: relative path\n", pattern)

		return new(Match) // for consistency
	}

	match := MatchModules(pattern)
	if len(match.Modules) == 0 {
		fmt.Fprintf(os.Stderr, "warning: %q matched no modules\n", pattern)
	}
	if len(match.Modules) > 1 {
		fmt.Fprintf(os.Stderr, "warning: %q matched multiple modules\n", pattern)
	}

	return match
}

// load loads the module at dirpath, that must be a directory containing the
// go.mod file.
func load(root, dirpath string) (*Module, error) {
	// TODO(mperillo): Should we check that the module path is valid (contains
	// a dot in the first path segment)?  See golang.org/x/mod/module#CheckPath
	path := filepath.Join(dirpath, "go.mod")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err // go.mod file was removed
	}

	// Construct the module.
	mod := &Module{
		Dir:   dirpath,
		GoMod: path,
		Main:  true,
	}
	file, err := modfile.Parse(path, data, nil) // should we use ParseLax?
	if err != nil {
		return nil, err
	}

	// Handle missing module and go directives.
	if file.Module != nil {
		mod.Path = file.Module.Mod.Path
		mod.Version = file.Module.Mod.Version
	} else {
		fmt.Fprintf(os.Stderr, "warning: missing module directive in %s\n", path)
		mod.Path, _ = filepath.Rel(root, dirpath) // it is safe to ignore err
	}
	if file.Go != nil {
		mod.GoMod = file.Go.Version
	} else {
		fmt.Fprintf(os.Stderr, "warning: missing go directive in %s\n", path)
		mod.GoMod = "1.13"
	}

	// The module path, as defined in the module directive in go.mod, must be
	// inside $GOPATH.
	if filepath.Join(root, mod.Path) != dirpath {
		// TODO(mperillo): Should we just print a warning in case mod.Path is
		// not in $GOPATH?
		return nil, fmt.Errorf("module %s: not in $GOPATH: %s", mod.Path, dirpath)
	}

	return mod, nil
}

func gopath() string {
	// TODO(mperillo): Should we get $GOPATH from go env, instead of os.Getenv?
	value, ok := os.LookupEnv("GOPATH")
	if !ok {
		panic("GOPATH not found")
	}

	return value
}

// isDir returns true if path exists and it is a directory.
func isDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fi.IsDir()
}

// isLiteral returns true if pattern does not contain wildcards.
func isLiteral(pattern string) bool {
	return !strings.Contains(pattern, "...")
}

// isRelativePath reports whether pattern should be interpreted as a directory
// path relative to the current directory, as opposed to a pattern matching
// module paths.
func isRelativePath(path string) bool {
	return path == "." || path == ".." ||
		strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../")
}

// matchPattern(pattern)(name) reports whether name matches pattern.
// Pattern is a limited glob pattern in which '...' has the same meaning as in
// the go tool.
func matchPattern(pattern string) func(name string) bool {
	// Convert pattern to regular expression.
	// The strategy for the trailing /... is the same as the one used in the go
	// tool.
	re := regexp.QuoteMeta(pattern)
	if strings.HasSuffix(re, `/\.\.\.`) {
		re = strings.TrimSuffix(re, `/\.\.\.`) + `(/\.\.\.)?`
	}
	re = strings.ReplaceAll(re, `\.\.\.`, `.*`)

	reg := regexp.MustCompile(`^` + re + `$`)

	return func(name string) bool {
		return reg.MatchString(name)
	}
}
