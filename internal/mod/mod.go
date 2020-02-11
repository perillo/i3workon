// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The code for module matching has been adapted from
// src/cmd/go/internal/search/search.go in the Go source distribution.
// Copyright 2017 The Go Authors. All rights reserved.

// Package mod implements support for local modules.
//
// A local module is a module whose module path, as defined in the module
// directive in go.mod, is inside $GOPATH.
package mod

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/perillo/gocmd/modlist"
)

// Module represents a local module.
type Module = modlist.Module

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
			mod, err := load(dir)
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
	if filepath.IsAbs(pattern) {
		return nil, fmt.Errorf("resolve: %q: absolute path", pattern)
	}
	if isRelativePath(pattern) {
		return nil, fmt.Errorf("resolve: %q: relative path", pattern)
	}

	m := MatchModules(pattern)
	switch {
	case len(m.Modules) == 0:
		return nil, fmt.Errorf("resolve: %q: no modules matched", pattern)
	case len(m.Modules) > 1:
		return nil, fmt.Errorf("resolve: %q: multiple modules matched", pattern)
	}

	return m.Modules[0], nil
}

// load loads the module at dirpath, that must be a directory containing the
// go.mod file.
// TODO(mperillo): Should load be reimplemented using
// https://godoc.org/golang.org/x/mod/modfile instead of go list -m?
func load(dirpath string) (*Module, error) {
	// TODO(mperillo): Should we check that the module path is valid (contains
	// a dot in the first path segment)?  See golang.org/x/mod/module#CheckPath
	l := modlist.Loader{
		Dir: dirpath,
	}
	mods, err := l.Load() // load current module
	if err != nil {
		return nil, err
	}
	mod := mods[0] // len(mods) is always > 0

	// The module path, as defined in the module directive in go.mod, must be
	// inside $GOPATH.
	found := false
	for _, root := range filepath.SplitList(gopath()) {
		modpath := filepath.Join(root, "src", mod.Path)
		if modpath == dirpath {
			found = true

			break
		}
	}
	if !found {
		return nil, fmt.Errorf("module %s: not in $GOPATH", mod.Path)
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
