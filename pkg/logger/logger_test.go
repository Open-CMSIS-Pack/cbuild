/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package logger

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomErrorMethod(t *testing.T) {
	// Set up buffer to capture log output
	var logOutput bytes.Buffer
	SetOutput(&logOutput)

	// Simulate an exec.ExitError
	cmd := exec.Command("cmd", "/C", "exit", "1")
	err := cmd.Run()

	// Log the error using the custom Error method
	if exitErr, ok := err.(*exec.ExitError); ok {
		Error(exitErr)
	}

	aa := logOutput.String()
	fmt.Println(aa)
	// Assert that custom logic was applied by checking the log output for "exit_code"
	assert.Contains(t, aa, "info cbuild: exit status 1\n")

	// Clear the buffer
	logOutput.Reset()

	// Test logging a generic error
	genericErr := errors.New("generic error")
	Error(genericErr)
	// Assert that the original logrus Error behavior was called
	assert.Contains(t, logOutput.String(), "generic error", "Expected log output to contain 'generic error'")
}
