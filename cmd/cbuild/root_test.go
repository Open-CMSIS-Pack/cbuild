/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func executeAndCheck(t *testing.T, cmd *cobra.Command, arguments []string) {
	cmd.SetArgs(arguments)
	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestCommands(t *testing.T) {
	cmd := NewRootCmd()

	executeAndCheck(t, cmd, []string{"-v"})

	// TODO: Implement tests
	//	executeAndCheck(t, cmd, []string{"../testdata/minimal.cprj",
	//	                                 "../../test/run/IntDir/",
	//									 "../../test/run/OutDir/"})
}
