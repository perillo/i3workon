// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The isRelativePath function has been adapted from
// src/go/build/build.go in the Go source distribution.
// Copyright 2011 The Go Authors. All rights reserved.

// Package mod implements support for local modules.
//
// A local module is a module whose module path, as defined in the module
// directive in go.mod, is inside $GOPATH.
package mod

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/perillo/gocmd/modlist"
)

// Module represents a local module.
type Module = modlist.Module

// Resolve resolves pattern to a local Go module.
//
// pattern can be an absolute directory or a module path.  A module path is
// resolved relative to $GOPATH.
func Resolve(pattern string) (*Module, error) {
	if isRelativePath(pattern) {
		// Reject it now, since it will be rejected later.
		return nil, fmt.Errorf("resolve: %q: relative path", pattern)
	}

	if filepath.IsAbs(pattern) {
		mod, err := load(pattern)
		if err != nil {
			return nil, fmt.Errorf("resolve: %q: %q", pattern, err)
		}

		return mod, nil
	}

	// Make the remote module path local, relative to $GOPATH.
	for _, root := range filepath.SplitList(gopath()) {
		path := filepath.Join(root, "src", pattern)
		if isDir(path) {
			mod, err := load(path)
			if err != nil {
				return nil, fmt.Errorf("resolve: %q: %q", pattern, err)
			}

			return mod, nil
		}
	}

	return nil, fmt.Errorf("resolve: %q: unable to resolve", pattern)
}

// load loads the module at dirpath.
// TODO(mperillo): Should load be reimplemented using
// https://godoc.org/golang.org/x/mod/modfile instead of go list -m?
func load(dirpath string) (*Module, error) {
	if !isDir(dirpath) {
		return nil, errors.New("not a directory")
	}

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

	// Unfortunately go list -m does not return an error if there is no go.mod
	// file.
	if mod.GoMod == "" {
		return nil, errors.New("not a module")
	}

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
		return nil, errors.New("module not in $GOPATH")
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

// isRelativePath reports whether path is relative, like ".", "..", "./foo", or
// "../foo".
func isRelativePath(path string) bool {
	return path == "." || path == ".." ||
		strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../")
}
