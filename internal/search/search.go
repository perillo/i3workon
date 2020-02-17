// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The code for module matching has been adapted from
// src/cmd/go/internal/search/search.go in the Go source distribution.
// Copyright 2017 The Go Authors. All rights reserved.

// Package search implements support for searching local modules.
//
// For the purpose of search, a local module is a module whose go.mod file is
// inside $GOPATH.
package search

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Module represents a local raw module.
type Module struct {
	Path  string // module path
	Dir   string // directory holding files for this module
	GoMod string // path to go.mod file for this module
	Root  string // Go path dir containing this module
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
			if filepath.Base(path) != "go.mod" {
				return nil
			}

			mod := mkmod(root, filepath.Dir(path))
			if !match(mod.Path) {
				return nil
			}
			m.Modules = append(m.Modules, mod)

			return nil
		})
	}

	return m
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
		fmt.Fprintf(os.Stderr, "warning: %q matched multiple modules:\n", pattern)
		for _, m := range match.Modules {
			fmt.Fprintf(os.Stderr, "\t%s\n", m.Path)
		}
	}

	return match
}

// mkmod synthesizes a raw module at dirpath, just for the sake of pattern
// matching.
func mkmod(root, dirpath string) *Module {
	mod := &Module{
		Dir:   dirpath,
		GoMod: filepath.Join(dirpath, "go.mod"),
		Root:  root,
	}
	mod.Path, _ = filepath.Rel(root, dirpath) // it is safe to ignore err

	return mod
}

func gopath() string {
	// TODO(mperillo): Should we get $GOPATH from go env, instead of os.Getenv?
	value, ok := os.LookupEnv("GOPATH")
	if !ok {
		panic("GOPATH not found")
	}

	return value
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
