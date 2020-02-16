// Copyright 2020 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The implementation of invoke is based on
// github.com/perillo/cmdgo/internal/invoke, but simplified to match the
// requirements of i3workon.

package i3

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// invoke invokes the i3-msg command with the specified msgtype and msg.
//
// If the i3-msg command returns a non 0 exit status, invoke will return a
// nil []byte and an error.
//
// If the i3-msg command returns a 0 exit status, invoke will return the
// stdout content as a []byte and a nil error.
//
// The child process stderr will be redirected to the parent process stderr.
func invoke(msgtype, msg string) ([]byte, error) {
	stdout := new(bytes.Buffer)

	cmd := exec.Command("i3-msg", "-t", msgtype, msg)
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("i3-msg -t %s %s: %v", msgtype, msg, err)
	}

	return stdout.Bytes(), nil
}
