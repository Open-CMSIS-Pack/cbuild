/*
 * Copyright (c) 2023-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	"path/filepath"
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild/v2/cmd/cbuild/commands"
	"github.com/stretchr/testify/assert"
)

func TestListToolchainsCommand(t *testing.T) {
	assert := assert.New(t)
	csolutionFile := filepath.Join(testRoot, testDir, "TestSolution/test.csolution.yml")

	t.Run("invalid flag", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "toolchains", "--invalid"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("multiple arguments", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "toolchains", csolutionFile, csolutionFile})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test list toolchain", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "toolchains", csolutionFile})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test help", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "toolchains", "-h"})
		err := cmd.Execute()
		assert.Nil(err)
	})
}
