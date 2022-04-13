/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testRoot = "../../test"

func TestGetExecutablePath(t *testing.T) {
	assert := assert.New(t)

	t.Run("get executable path", func(t *testing.T) {
		_, err := GetExecutablePath()
		assert.Nil(err)
	})
}

func TestUpdateEnvVars(t *testing.T) {
	assert := assert.New(t)

	t.Run("test update environment variables", func(t *testing.T) {
		binPath := testRoot + "/bin"
		etcPath := testRoot + "/etc"
		env := UpdateEnvVars(binPath, etcPath)
		binPath, _ = filepath.Abs(binPath)
		etcPath, _ = filepath.Abs(etcPath)
		assert.Equal(env.BuildRoot, binPath)
		assert.Equal(env.CompilerRoot, etcPath)
		assert.NotEmpty(env.PackRoot)
	})

	t.Run("test update environment variables, with CMSIS_PACK_ROOT", func(t *testing.T) {
		binPath := testRoot + "/bin"
		etcPath := testRoot + "/etc"
		packRoot, _ := filepath.Abs(testRoot + "/packs")
		_ = os.Setenv("CMSIS_PACK_ROOT", packRoot)
		env := UpdateEnvVars(binPath, etcPath)
		binPath, _ = filepath.Abs(binPath)
		etcPath, _ = filepath.Abs(etcPath)
		assert.Equal(env.BuildRoot, binPath)
		assert.Equal(env.CompilerRoot, etcPath)
		assert.NotEmpty(env.PackRoot)
	})
}

func TestGetDefaultCmsisPackRoot(t *testing.T) {
	assert := assert.New(t)

	t.Run("get default cmsis pack root", func(t *testing.T) {
		root := GetDefaultCmsisPackRoot()
		assert.NotEmpty(root)
	})
}
