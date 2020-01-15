// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command workon starts the preferred terminal with its current working
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

const terminal = "urxvt"

var editor = flag.String("editor", defaultEditor(), "preferred editor to use")

func defaultEditor() string {
	if editor, ok := os.LookupEnv("VISUAL"); ok {
		return editor // even if it is empty
	}
	if editor, ok := os.LookupEnv("EDITOR"); ok {
		return editor // even if it is empty
	}

	// TODO(mperillo): Use a suitable editor based on the operating system.
	return ""
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()

		os.Exit(2)
	}
	path := flag.Arg(0)

	if err := startTerminal(path); err != nil {
		log.Fatal(err)
	}
	if err := startEditor(path, *editor); err != nil {
		log.Fatal(err)
	}
}

// startTerminal starts the preferred terminal and sets its current working
// directory to dirpath.
func startTerminal(dirpath string) error {
	attr := os.ProcAttr{
		Dir: dirpath,
		Env: os.Environ(),
	}
	args := []string{"-cd", dirpath}
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
	fmt.Println(files)

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
