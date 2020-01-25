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
)

var (
	workspace = flag.String("workspace", "", "workspace where to switch to")
	terminal  = flag.String("terminal", defaultTerminal(), "terminal to use")
	editor    = flag.String("editor", defaultEditor(), "preferred editor to use")
)

func defaultTerminal() string {
	// NOTE(mperillo): The TERM environment variable is not usable, so we use
	// the TERMINAL environment variable instead, as it is done in
	// i3-sensible-terminal.
	if term, ok := os.LookupEnv("TERMINAL"); ok {
		return term // even it is is empty
	}

	return "i3-sensible-terminal"
}

func defaultEditor() string {
	if editor, ok := os.LookupEnv("VISUAL"); ok {
		return editor // even if it is empty
	}
	if editor, ok := os.LookupEnv("EDITOR"); ok {
		return editor // even if it is empty
	}

	return "i3-sensible-editor"
}

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
	if *terminal == "" {
		log.Fatal("no terminal emulator available")
	}
	if *editor == "" {
		log.Fatal("no editor available")
	}

	// Validate the argument.
	path := flag.Arg(0)
	switch t, err := isDir(path); {
	case err != nil:
		log.Fatal(err)
	case !t:
		log.Fatalf("path %s is not a directory", path)
	}

	if *workspace != "" {
		if err := switchToWorkspace(*workspace); err != nil {
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

// isDir returns true if path exists and it is a directory.
func isDir(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fi.IsDir(), nil
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
