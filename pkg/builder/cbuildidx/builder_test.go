/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cbuildidx

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	builder "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/inittest"
	utils "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../../test"

func init() {
	inittest.TestInitialization(testRoot)
}

type RunnerMock struct{}

func (r RunnerMock) ExecuteCommand(program string, quiet bool, args ...string) (string, error) {
	if strings.Contains(program, "cbuild2cmake") {
		_ = os.MkdirAll(testRoot+"/run/tmp", 0755)
		cmakelistFile := testRoot + "/run/tmp/CMakeLists.txt"
		file, _ := os.Create(cmakelistFile)
		defer file.Close()
	} else if strings.Contains(program, "cpackget") {
	} else if strings.Contains(program, "cmake") {
	} else if strings.Contains(program, "ninja") {
	} else if strings.Contains(program, "xmllint") {
	} else {
		return "", errors.New("invalid command")
	}
	return "", nil
}

func TestCheckCbuildIdx(t *testing.T) {
	assert := assert.New(t)

	b := CbuildIdxBuilder{
		builder.BuilderParams{
			Runner: RunnerMock{},
		},
	}

	t.Run("test valid cprj", func(t *testing.T) {
		b.InputFile = testRoot + "/run/Hello.cbuild-idx.yml"
		err := b.checkCbuildIdx()
		assert.Nil(err)
	})

	t.Run("test existent file, invalid extension", func(t *testing.T) {
		b.InputFile = testRoot + "/run/main.c"
		err := b.checkCbuildIdx()
		assert.Error(err)
	})

	t.Run("test invalid file", func(t *testing.T) {
		b.InputFile = testRoot + "/run/invalid-file.cbuild-idx.yml"
		err := b.checkCbuildIdx()
		assert.Error(err)
	})
}

func TestGetDirs(t *testing.T) {
	assert := assert.New(t)

	b := CbuildIdxBuilder{
		builder.BuilderParams{
			Runner: RunnerMock{},
		},
	}

	b.InputFile = testRoot + "/run/Hello.cbuild-idx.yml"
	b.Options.Contexts = []string{"Hello.Debug+AVH"}

	t.Run("test valid directories in cprj", func(t *testing.T) {
		dirs, err := b.getDirs()
		assert.Nil(err)
		intDir, _ := filepath.Abs(testRoot + "/run/tmp")
		outDir, _ := filepath.Abs(testRoot + "/run/out/AVH")
		assert.Equal(intDir, dirs.IntDir)
		assert.Equal(outDir, dirs.OutDir)
	})

	t.Run("test valid directories as arguments", func(t *testing.T) {
		b.Options.IntDir = "cmdOptionsIntDir"
		b.Options.OutDir = "cmdOptionsOutDir"
		dirs, err := b.getDirs()
		assert.Nil(err)
		intDir, _ := filepath.Abs(testRoot + "/run/tmp")
		outDir, _ := filepath.Abs(b.Options.OutDir)
		assert.Equal(intDir, dirs.IntDir)
		assert.Equal(outDir, dirs.OutDir)
	})

	t.Run("test invalid cprj", func(t *testing.T) {
		b.InputFile = testRoot + "/run/invalid-file.cbuild-idx.yml"
		_, err := b.getDirs()
		assert.Error(err)
	})
}

func TestClean(t *testing.T) {
	assert := assert.New(t)

	b := CbuildIdxBuilder{
		builder.BuilderParams{
			Runner: RunnerMock{},
		},
	}
	var dirs builder.BuildDirs
	var vars builder.InternalVars

	t.Run("test clean directories, invalid tool", func(t *testing.T) {
		vars.CmakeBin = testRoot + "/bin/invalid-tool"

		dirs.OutDir = testRoot + "/run/OutDir"
		_ = os.MkdirAll(dirs.OutDir, 0755)
		err := b.clean(dirs, vars)
		assert.Error(err)

		dirs.IntDir = testRoot + "/run/IntDir"
		_ = os.MkdirAll(dirs.IntDir, 0755)
		err = b.clean(dirs, vars)
		assert.Error(err)
	})

	t.Run("test clean directories", func(t *testing.T) {
		vars.CmakeBin = testRoot + "/bin/cmake"
		dirs.IntDir = testRoot + "/run/tmp"
		dirs.OutDir = testRoot + "/run/OutDir"
		_ = os.MkdirAll(dirs.IntDir, 0755)
		_ = os.MkdirAll(dirs.OutDir, 0755)
		err := b.clean(dirs, vars)
		assert.Nil(err)
	})
}

func TestBuild(t *testing.T) {
	assert := assert.New(t)
	configs := inittest.GetTestConfigs(testRoot)

	b := CbuildIdxBuilder{
		builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: testRoot + "/run/Hello.cbuild-idx.yml",
			Options: builder.Options{
				OutDir: testRoot + "/run/OutDir",
				Packs:  true,
			},
			InstallConfigs: utils.Configurations{
				BinPath: configs.BinPath,
				BinExtn: configs.BinExtn,
				EtcPath: configs.EtcPath,
			},
		},
	}
	b.Options.Contexts = []string{"Hello.Debug+AVH"}

	t.Run("test build cbuild-idx", func(t *testing.T) {
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test build cprj quiet", func(t *testing.T) {
		b.Options.Quiet = true
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test build cprj debug", func(t *testing.T) {
		b.Options.Debug = true
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test rebuild cprj", func(t *testing.T) {
		b.Options.Rebuild = true
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test build log", func(t *testing.T) {
		b.Options.LogFile = testRoot + "/run/log/test.log"
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test build jobs", func(t *testing.T) {
		b.Options.Jobs = 1
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test build verbose", func(t *testing.T) {
		b.Options.Debug = false
		b.Options.Verbose = true
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test build makefile generator", func(t *testing.T) {
		b.Options.OutDir = testRoot + "/run/OutDir"
		b.Options.Debug = true
		b.Options.Generator = "Unix Makefiles"
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test clean cprj", func(t *testing.T) {
		b.Options.Clean = true
		err := b.Build()
		assert.Nil(err)
	})
}
