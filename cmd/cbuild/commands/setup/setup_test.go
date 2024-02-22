/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package setup_test

import (
	"path/filepath"
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild/v2/cmd/cbuild/commands"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../../../test"
const testDir = "command"

func TestSetupCommand(t *testing.T) {
	assert := assert.New(t)
	csolutionFile := filepath.Join(testRoot, testDir, "TestSolution/test.csolution.yml")

	t.Run("test valid command", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"setup", csolutionFile})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("No arguments", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"setup"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("invalid flag", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"setup", csolutionFile, "--invalid"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("multiple arguments", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"setup", csolutionFile, csolutionFile})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test setup help", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"setup", "-h"})
		err := cmd.Execute()
		assert.Nil(err)
	})

	t.Run("test setup invalid input argument", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"setup", "test.cbuild.yml"})
		err := cmd.Execute()
		assert.Nil(err)
	})
}
