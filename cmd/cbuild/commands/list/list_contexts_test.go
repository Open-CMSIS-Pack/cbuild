/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	"cbuild/cmd/cbuild/commands"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListContextsCommand(t *testing.T) {
	assert := assert.New(t)
	csolutionFile := testRoot + "/run/TestSolution/test.csolution.yml"

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
