// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package i3 provides support for the i3 wm.
package i3

import (
	"fmt"
)

// Workspace switches to workspace.
func Workspace(workspace string) error {
	// With i3, workspace can be an integer or a generic string.
	msg := "workspace" + " " + workspace
	_, err := invoke("command", msg)
	if err != nil {
		return fmt.Errorf("workspace %s: %w", workspace, err)
	}

	return nil
}
