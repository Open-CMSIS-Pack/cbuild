/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	"cbuild/cmd/cbuild/commands"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListEnvironmentCommand(t *testing.T) {
	assert := assert.New(t)
	os.Setenv("CMSIS_BUILD_ROOT", testRoot+"/run/bin")

	t.Run("invalid args", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "environment", "--invalid"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test list environment", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "environment"})
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
