// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The code and the API has been adapted from
// src/cmd/go/internal/load/pkg.go in the Go source distribution.
// Copyright 2011 The Go Authors. All rights reserved.

// Package load loads local modules.
//
// A local module is a module whose module path, as defined in the module
// directive in go.mod, is inside $GOPATH.
package load

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/perillo/i3workon/internal/search"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

// Module represents a local module.
type Module struct {
	Path      string       // module path
	Version   string       // module version, always empty
	Main      bool         // is this the main module?
	Dir       string       // directory holding files for this module
	GoMod     string       // path to go.mod file for this module
	GoVersion string       // go version used in module
	Root      string       // Go path dir containing this module
	Error     *ModuleError // error loading module
}

// Name returns the short name of the module.
func (m *Module) Name() string {
	// The code assumes that Path is correct, since it was validate by load.
	i := strings.LastIndexByte(m.Path, '/')
	if i > 0 {
		return m.Path[i+1:]
	}

	return m.Path
}

// ModuleError represents a module error.
type ModuleError struct {
	Err string // the error itself
}

// Error implements the error interface.
func (me *ModuleError) Error() string {
	return me.Err
}

// Modules returns the modules named by the command line argument arg. If a
// named module cannot be loaded at all (for example, if the go.mod file has
// problems), then Modules prints an error and does not include that module in
// the results.
func Modules(arg string) []*Module {
	var mods []*Module
	for _, mod := range ModulesAndErrors(arg) {
		if mod.Error != nil {
			fmt.Fprintf(os.Stderr, "can't load module: %s\n", mod.Error)

			continue
		}
		mods = append(mods, mod)
	}

	return mods
}

// ModulesAndErrors is like Modules but returns a *Module for every argument,
// even the ones that cannot be loaded at all.
// The modules that fail to load will have m.Error != nil.
func ModulesAndErrors(pattern string) []*Module {
	match := search.ModulePath(pattern)
	var mods []*Module

	for _, mod := range match.Modules {
		m, err := load(mod)
		if err != nil {
			err := &ModuleError{
				Err: err.Error(),
			}
			m.Error = err
		}
		mods = append(mods, m)
	}

	return mods
}

// load loads the raw module m.
func load(raw *search.Module) (*Module, error) {
	// TODO(mperillo): Should we check that the module path is valid (contains
	// a dot in the first path segment)?  See golang.org/x/mod/module#CheckPath
	data, err := ioutil.ReadFile(raw.GoMod)
	if err != nil {
		return nil, err // go.mod file was removed
	}

	// Construct the module.  The Main field is set to false, since it is
	// onerous to determine if the module is a main module.
	// See https://tip.golang.org/ref/modules#glos-main-module.
	mod := &Module{
		Path:  raw.Path,
		Main:  false,
		Dir:   raw.Dir,
		GoMod: raw.GoMod,
		Root:  raw.Root,
	}
	file, err := modfile.Parse(raw.GoMod, data, nil) // should we use ParseLax?
	if err != nil {
		return nil, err
	}

	// Handle missing module directive.
	if file.Module != nil {
		modpath := file.Module.Mod.Path
		// We should probably use CheckPath, but since go mod init will not do
		// it, we do the same.
		if err := module.CheckImportPath(modpath); err != nil {
			err := &ModuleError{
				Err: fmt.Sprintf("module %s: invalid importpath: %s", modpath, err),
			}
			mod.Error = err
		}

		mod.Path = modpath
		mod.Version = file.Module.Mod.Version
	} else {
		fmt.Fprintf(os.Stderr, "warning: missing module directive in %s\n", raw.GoMod)
	}

	// Handle missing go directive.
	if file.Go != nil {
		mod.GoMod = file.Go.Version
	} else {
		fmt.Fprintf(os.Stderr, "warning: missing go directive in %s\n", raw.GoMod)
		mod.GoMod = "1.13"
	}

	// The module path, as defined in the module directive in go.mod, must be
	// inside $GOPATH.
	if filepath.Join(mod.Root, mod.Path) != mod.Dir {
		err := &ModuleError{
			Err: fmt.Sprintf("module %s: not in $GOPATH: %s", mod.Path, mod.Dir),
		}
		mod.Error = err
	}

	return mod, nil
}
