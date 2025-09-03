/*
 * Copyright (c) 2024-2025 Arm Limited. All rights reserved.
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

	t.Run("test valid command with -a", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"setup", csolutionFile, "--active", "test"})

		err := cmd.Execute()

		// Though the command is valid, It fails for other reasons
		assert.Error(err)
		assert.Contains(err.Error(), "couldn't locate '../etc' directory relative to")
	})

	t.Run("test valid command with -a <empty arg>", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"setup", csolutionFile, "--active", ""})

		err := cmd.Execute()

		// Though the command is valid, It fails for other reasons
		assert.Error(err)
		assert.Contains(err.Error(), "couldn't locate '../etc' directory relative to")
	})

	t.Run("test invalid arguments to -a option", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		args := []string{"setup", csolutionFile, "-a", "-d"}
		cmd.SetArgs(args)

		err := cmd.Execute()
		assert.Error(err)
		assert.EqualError(err, "invalid input argument for '-a'")
	})

	t.Run("test invalid command with -a and -S", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		args := []string{"setup", csolutionFile, "-a", "test", "-S"}
		cmd.SetArgs(args)

		err := cmd.Execute()
		assert.Error(err)
		assert.EqualError(err, "invalid command line arguments. Options '-a' and '-S' are mutually exclusive")
	})

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
