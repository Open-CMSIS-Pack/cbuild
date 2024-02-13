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

func TestListContextsCommand(t *testing.T) {
	assert := assert.New(t)
	csolutionFile := filepath.Join(testRoot, testDir, "TestSolution/test.csolution.yml")

	t.Run("No arguments", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "contexts"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("invalid flag", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "contexts", "--invalid"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("multiple arguments", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "contexts", csolutionFile, csolutionFile})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test list contexts", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "contexts", csolutionFile})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test list context help", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "contexts", "-h"})
		err := cmd.Execute()
		assert.Nil(err)
	})
}
