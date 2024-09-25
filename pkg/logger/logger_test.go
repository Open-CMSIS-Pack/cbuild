/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package logger

import (
	"bytes"
	"errors"
	"os/exec"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomErrorMethod(t *testing.T) {
	// Set up buffer to capture log output
	var logOutput bytes.Buffer
	SetOutput(&logOutput)

	//TEST 1: Simulate an exec.ExitError
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows: using cmd to force exit with non-zero status
		cmd = exec.Command("cmd", "/C", "exit", "1")
	} else {
		// Unix-based (Linux/macOS): using sh to force exit with non-zero status
		cmd = exec.Command("sh", "-c", "exit 1")
	}
	err := cmd.Run()
	// Log the error using the custom Error method
	if exitErr, ok := err.(*exec.ExitError); ok {
		Error(exitErr)
	}
	// Assert that custom logic was applied by checking the log output for "exit status"
	assert.Contains(t, logOutput.String(), "info cbuild: exit status 1\n")

	// Clear the buffer
	logOutput.Reset()

	//TEST 2: Test logging a generic error
	genericErr := errors.New("generic error")
	Error(genericErr)
	// Assert that the original logrus Error behavior was called
	assert.Contains(t, logOutput.String(), "generic error", "Expected log output to contain 'generic error'")
}
