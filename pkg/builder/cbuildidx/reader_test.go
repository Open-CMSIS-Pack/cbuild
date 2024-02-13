/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cbuildidx

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBuildDirs(t *testing.T) {
	assert := assert.New(t)
	t.Run("test get build directories when file is missing", func(t *testing.T) {
		cbuildFile := filepath.Join(testRoot, testDir, "missing.cbuild.yml")
		_, _, err := GetBuildDirs(cbuildFile)
		assert.Error(err)
	})

	t.Run("test get build directories from .cbuild.yml", func(t *testing.T) {
		cbuildFile := filepath.Join(testRoot, testDir, "Hello.Debug+AVH.cbuild.yml")
		intDir, outDir, err := GetBuildDirs(cbuildFile)
		assert.Nil(err)
		assert.Equal("tmp/Hello/AVH/Debug", intDir)
		assert.Equal("out/AVH", outDir)
	})
}
