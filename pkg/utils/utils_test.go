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

func TestParseContext(t *testing.T) {
	assert := assert.New(t)

	t.Run("test empty context", func(t *testing.T) {
		_, err := ParseContext("")
		assert.Error(err)
	})

	t.Run("test invalid context 1", func(t *testing.T) {
		_, err := ParseContext("ProjName+Target.Build")
		assert.Error(err)
	})

	t.Run("test invalid context 2", func(t *testing.T) {
		_, err := ParseContext(".Build+Target")
		assert.Error(err)
	})

	t.Run("test valid context", func(t *testing.T) {
		contextItem, err := ParseContext("ProjName.Build+Target")
		assert.Nil(err)
		assert.Equal(contextItem.ProjectName, "ProjName")
		assert.Equal(contextItem.BuildType, "Build")
		assert.Equal(contextItem.TargetType, "Target")
	})
}

func TestParseCbuildIndexFile(t *testing.T) {
	assert := assert.New(t)

	t.Run("test file not available", func(t *testing.T) {
		_, err := ParseCbuildIndexFile("Unknown.cbuild-idx.yml")
		assert.Error(err)
	})

	t.Run("test", func(t *testing.T) {
		data, err := ParseCbuildIndexFile(testRoot + "/run/Test.cbuild-idx.yml")
		assert.Nil(err)
		assert.Equal(data.BuildIdx.GeneratedBy, "csolution 1.4.0")
		assert.Equal(data.BuildIdx.Cdefault, "HelloWorld.cdefault.yml")
		assert.Equal(data.BuildIdx.Csolution, "HelloWorld.csolution.yml")
		assert.Equal(len(data.BuildIdx.Cprojects), 2)
		assert.Equal(data.BuildIdx.Cprojects[0].Cproject, "cm0plus/HelloWorld_cm0plus.cproject.yml")
		assert.Equal(data.BuildIdx.Cprojects[1].Cproject, "cm4/HelloWorld_cm4.cproject.yml")
		assert.Equal(data.BuildIdx.Licenses, "test123")
		assert.Equal(len(data.BuildIdx.Cbuilds), 4)
		assert.Equal(data.BuildIdx.Cbuilds[0].Cbuild, "cm0plus/HelloWorld_cm0plus.Debug+FRDM-K32L3A6.cbuild.yml")
		assert.Equal(data.BuildIdx.Cbuilds[1].Cbuild, "cm0plus/HelloWorld_cm0plus.Release+FRDM-K32L3A6.cbuild.yml")
		assert.Equal(data.BuildIdx.Cbuilds[2].Cbuild, "cm4/HelloWorld_cm4.Debug+FRDM-K32L3A6.cbuild.yml")
		assert.Equal(data.BuildIdx.Cbuilds[3].Cbuild, "cm4/HelloWorld_cm4.Release+FRDM-K32L3A6.cbuild.yml")
	})
}
