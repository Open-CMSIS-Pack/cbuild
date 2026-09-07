/*
 * Copyright (c) 2025 Arm Limited. All rights reserved.
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

func TestListTargetSetsCommand(t *testing.T) {
	assert := assert.New(t)
	csolutionFile := filepath.Join(testRoot, testDir, "TestSolution/test.csolution.yml")

	t.Run("No arguments", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "targets"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("invalid flag", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "targets", "--invalid"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("multiple arguments", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "targets", csolutionFile, csolutionFile})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test list targets", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "targets", csolutionFile})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test list targets help", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "targets", "-h"})
		err := cmd.Execute()
		assert.Nil(err)
	})

	t.Run("test list target-sets alias", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "target-sets", "-h"})
		err := cmd.Execute()
		assert.Nil(err)
	})
}
