/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteCommand(t *testing.T) {
	assert := assert.New(t)
	runner := Runner{}
	t.Run("execute command normal verbosity", func(t *testing.T) {
		err := runner.ExecuteCommand("go", false, "version")
		assert.Nil(err)
	})

	t.Run("execute command quiet", func(t *testing.T) {
		err := runner.ExecuteCommand("go", true, "version")
		assert.Nil(err)
	})
}
