// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package i3 provides support for the i3 wm.
package i3

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// workspace represents an i3 workspace.
// See https://i3wm.org/docs/ipc.html#_workspaces_reply.
type workspace struct {
	// Num is the logical number of the workspace.  For named workspaces, Num
	// will be -1.
	Num int
	// Name is the name of this workspace (by default num+1).
	Name string
	// Visible reports whether this workspace is currently visible on an output
	// (multiple workspaces can be visible at the same time).
	Visible bool
	// Focused reports whether this workspace currently has the focus (only one
	// workspace can have the focus at the same time).
	Focused bool
}

// Number returns the effective workspace number.
func (w *workspace) Number() int {
	// We only support the format "[n]" or "[n][:][NAME]", with n > 0.
	// Note that i3 is more flexible, since it will also accept "[n][NAME]".
	// See https://i3wm.org/docs/userguide.html#_strip_workspace_numbers_name.
	if num := parse(w.Name); num > 0 {
		return num
	}

	i := strings.IndexByte(w.Name, ':')
	if i < 0 {
		return 0
	}

	return parse(w.Name[:i])
}

// Workspace switches to workspace with specified number and name.
func Workspace(num int, name string) error {
	// With i3, workspace can be an integer or a generic string.
	msg := fmt.Sprintf("workspace %d:%s", num, name)
	_, err := invoke("command", msg)
	if err != nil {
		return fmt.Errorf("workspace %d:%s: %w", num, name, err)
	}

	return nil
}

// NextWorkspace returns the next available workspace number.
func NextWorkspace() (int, error) {
	list, err := workspaces()
	if err != nil {
		return 0, fmt.Errorf("next workspace: %w", err)
	}

	n := 1 // by default workspace numbers start at 1
	for _, w := range list {
		num := w.Number()
		if num == 0 {
			continue
		}
		if num > n {
			return n, nil
		}

		n++
	}

	return n, nil
}

// workspaces return the list of current workspaces.
func workspaces() ([]*workspace, error) {
	stdout, err := invoke("get_workspaces", "")
	if err != nil {
		return nil, fmt.Errorf("workspaces: %w", err)
	}

	list, err := decode(stdout)
	if err != nil {
		return nil, fmt.Errorf("workspaces: %w", err)
	}

	return list, nil
}

func decode(data []byte) ([]*workspace, error) {
	list := make([]*workspace, 0, 10)
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("JSON decode: %w", err)
	}

	return list, nil
}

// parse parses the string and returns a positive number, or 0 in case of
// errors.
func parse(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0
	}

	return n
}
