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

	"github.com/perillo/gocmd/pkglist"

	"github.com/perillo/i3workon/internal/i3"
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

	if *workspace == 0 {
		*workspace, err = i3.NextWorkspace()
		if err != nil {
			log.Fatal(err)
		}
	}
	if *workspace > 0 {
		// Set the workspace label to the module name.  Note that by default
		// workspace numbers start at 1.
		if err := i3.Workspace(*workspace, mod.Name()); err != nil {
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
	l := pkglist.Loader{
		Dir: dirpath,
	}

	list, err := l.Load("./...")
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, 10)
	for _, p := range list {
		for _, path := range concat(p.GoFiles, p.CgoFiles, p.IgnoredGoFiles) {
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

// concat concatenates args into a single []string.
func concat(args ...[]string) []string {
	var buf []string
	for _, arg := range args {
		buf = append(buf, arg...)
	}

	return buf
}
