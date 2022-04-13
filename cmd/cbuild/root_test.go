/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"os"
	"testing"

	cp "github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../test"

type RunnerMock struct{}

func (r RunnerMock) ExecuteCommand(program string, quiet bool, args ...string) error {
	return nil
}

func init() {
	// Prepare test data
	_ = os.RemoveAll(testRoot + "/run")
	_ = cp.Copy(testRoot+"/data", testRoot+"/run")
}

func TestCommands(t *testing.T) {
	assert := assert.New(t)
	cprjFile := testRoot + "/run/minimal.cprj"

	t.Run("test version", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{"--version"})
		err := cmd.Execute()
		assert.Nil(err)
	})

	t.Run("test help", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{"--help"})
		err := cmd.Execute()
		assert.Nil(err)
	})

	t.Run("invalid flag", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{"--invalid"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("multiple arguments", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{cprjFile, cprjFile})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test build", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{cprjFile})
		err := cmd.Execute()
		assert.Error(err)
	})
}
