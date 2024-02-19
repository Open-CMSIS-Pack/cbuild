/*
 * Copyright (c) 2023-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild/v2/cmd/cbuild/commands"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../../../test"
const testDir = "command"

func TestListCommand(t *testing.T) {
	assert := assert.New(t)

	t.Run("No arguments", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list"})
		err := cmd.Execute()
		assert.Nil(err)
	})

	t.Run("invalid flag", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "--invalid"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("multiple arguments", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "test1", "test2"})
		err := cmd.Execute()
		assert.Nil(err)
	})

	t.Run("test list help", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "-h"})
		err := cmd.Execute()
		assert.Nil(err)
	})

	t.Run("test list invalid command", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "test"})
		err := cmd.Execute()
		assert.Nil(err)
	})
}
