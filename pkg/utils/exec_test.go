/*
 * Copyright (c) 2022-2026 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteCommand(t *testing.T) {
	assert := assert.New(t)
	runner := Runner{}
	t.Run("execute command normal verbosity", func(t *testing.T) {
		version, err := runner.ExecuteCommand("go", false, "version")
		assert.Nil(err)
		assert.Regexp("(go\\sversion\\sgo([\\d.]+).*)", version)
	})

	t.Run("execute command quiet", func(t *testing.T) {
		_, err := runner.ExecuteCommand("go", true, "version")
		assert.Nil(err)
	})

	t.Run("execute command from terminal", func(t *testing.T) {
		// Simulate terminal by overriding isTerminal function
		isTerminal = func() bool { return true }
		_, err := runner.ExecuteCommand("go", false, "version")
		assert.Nil(err)
	})
}

func TestExecuteCommandEx(t *testing.T) {
	assert := assert.New(t)
	t.Run("execute command normal verbosity", func(t *testing.T) {
		outStr, errStr, err := ExecuteCommand("go", "version")
		assert.Nil(err)
		assert.Empty(errStr)
		assert.Regexp("(go\\sversion\\sgo([\\d.]+).*)", outStr)
	})

	t.Run("execute invalid command", func(t *testing.T) {
		outStr, errStr, err := ExecuteCommand("go", "invalid")
		assert.Error(err)
		assert.Empty(outStr)
		assert.Equal("go invalid: unknown command\nRun 'go help' for usage.\n", errStr)
	})
}
