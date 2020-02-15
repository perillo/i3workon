// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command i3workon starts the preferred terminal with its current working
// directory set to the specified Go project and opens all the Go files in the
// preferred editor.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/tools/go/packages"

	"github.com/perillo/i3workon/internal/load"
)

var (
	workspace = flag.Int("workspace", 0, "workspace where to switch to")
	terminal  = flag.String("terminal", "i3-sensible-terminal", "terminal to use")
	editor    = flag.String("editor", "i3-sensible-editor", "editor to use")
)

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, "Usage: i3workon [flags] path")
		fmt.Fprintf(w, "Flags:\n")
		flag.PrintDefaults()
	}

	// Parse and validate the flags.
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()

		os.Exit(2)
	}

	// Resolve the pattern passed as argument.
	mod, err := resolve(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	path := mod.Dir

	if *workspace > 0 {
		// Set the workspace label to the module name.  Note that by default
		// workspace numbers start at 1.
		spec := fmt.Sprintf("%d:%s", *workspace, mod.Name())
		if err := switchToWorkspace(spec); err != nil {
			log.Fatal(err)
		}
	}
	if err := startTerminal(path, *terminal); err != nil {
		log.Fatal(err)
	}
	if err := startEditor(path, *editor); err != nil {
		log.Fatal(err)
	}
}

// switchToWorkspace switches to workspace.
func switchToWorkspace(workspace string) error {
	// With i3, workspace can be an integer or a generic string.
	msg := "workspace" + " " + workspace
	attr := os.ProcAttr{}
	args := []string{msg}
	proc, err := spawn("i3-msg", args, &attr)
	if err != nil {
		return fmt.Errorf("switching to workspace %s: %w", workspace, err)
	}

	// Detach the new process from the current process.
	if err := proc.Release(); err != nil {
		return fmt.Errorf("releasing i3-msg process: %w", err)
	}

	return nil
}

// startTerminal starts the preferred terminal and sets its current working
// directory to dirpath.
func startTerminal(dirpath, terminal string) error {
	attr := os.ProcAttr{
		Dir: dirpath,
		Env: os.Environ(),
	}
	args := []string{}
	proc, err := spawn(terminal, args, &attr)
	if err != nil {
		return fmt.Errorf("starting terminal: %w", err)
	}

	// Detach the new process from the current process.
	if err := proc.Release(); err != nil {
		return fmt.Errorf("releasing terminal process: %w", err)
	}

	return nil
}

// startEditor starts the preferred editor, sets its current working directory
// to dirpath and opens all the Go files in dirpath, including nested packages.
func startEditor(dirpath, editor string) error {
	files, err := gofiles(dirpath)
	if err != nil {
		return fmt.Errorf("finding files to edit: %w", err)
	}

	attr := os.ProcAttr{
		Dir: dirpath,
		Env: os.Environ(),
	}
	proc, err := spawn(editor, files, &attr)
	if err != nil {
		return fmt.Errorf("starting editor: %w", err)
	}

	// Detach the new process from the current process.
	if err := proc.Release(); err != nil {
		return fmt.Errorf("releasing editor process: %w", err)
	}

	return nil
}

// gofiles returns all the files in the Go project at dirpath.
func gofiles(dirpath string) ([]string, error) {
	cfg := &packages.Config{
		Dir:  dirpath,
		Mode: packages.NeedName | packages.NeedFiles,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("loading packages: %w", err)
	}
	if n := packages.PrintErrors(pkgs); n > 0 {
		return nil, fmt.Errorf("unable to correctly load %d packages", n)
	}

	files := make([]string, 0, 10)
	for _, pkg := range pkgs {
		for _, path := range pkg.GoFiles {
			// path is an absolute path, but since the editor working directory
			// has been set to dirpath, make it relative to it.
			path, err := filepath.Rel(dirpath, path)
			if err != nil {
				return nil, fmt.Errorf("processing go files: %v", err)
			}
			files = append(files, path)
		}
	}

	return files, nil
}

// spawn is a wrapper around os.StartProcess that ensures the first argument is
// set correctly.
func spawn(path string, argv []string, attr *os.ProcAttr) (*os.Process, error) {
	path, err := exec.LookPath(path)
	if err != nil {
		return nil, err
	}
	argv = append([]string{path}, argv...)

	return os.StartProcess(path, argv, attr)
}

// resolve resolves pattern to a local Go module.
//
// pattern can be not be an absolute or relative path.  A module path is
// resolved relative to $GOPATH.
func resolve(pattern string) (*load.Module, error) {
	mods := load.Modules(pattern)
	if len(mods) != 1 {
		return nil, fmt.Errorf("resolve %q: unable to resolve", pattern)
	}

	return mods[0], nil
}
