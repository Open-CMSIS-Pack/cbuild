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
const testDir = "cbuildidx"

func init() {
	inittest.TestInitialization(testRoot, testDir)
}

type RunnerMock struct{}

func (r RunnerMock) ExecuteCommand(program string, quiet bool, args ...string) (string, error) {
	if strings.Contains(program, "cbuild2cmake") {
		_ = os.MkdirAll(filepath.Join(testRoot, testDir, "tmp"), 0755)
		cmakelistFile := filepath.Join(testRoot, testDir, "tmp/CMakeLists.txt")
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

	t.Run("test valid cbuild-idx", func(t *testing.T) {
		b.InputFile = filepath.Join(testRoot, testDir, "Hello.cbuild-idx.yml")
		err := b.checkCbuildIdx()
		assert.Nil(err)
	})

	t.Run("test existent file, invalid extension", func(t *testing.T) {
		b.InputFile = filepath.Join(testRoot, testDir, "main.c")
		err := b.checkCbuildIdx()
		assert.Error(err)
	})

	t.Run("test invalid file", func(t *testing.T) {
		b.InputFile = filepath.Join(testRoot, testDir, "invalid-file.cbuild-idx.yml")
		err := b.checkCbuildIdx()
		assert.Error(err)
	})
}

func TestGetDirs(t *testing.T) {
	assert := assert.New(t)

	b := CbuildIdxBuilder{
		builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: filepath.Join(testRoot, testDir, "Hello.cbuild-idx.yml"),
		},
	}

	t.Run("test valid directories", func(t *testing.T) {
		dirs, err := b.getDirs("Hello.Debug+AVH")
		assert.Nil(err)
		intDir, _ := filepath.Abs(filepath.Join(testRoot, testDir, "tmp"))
		outDir, _ := filepath.Abs(filepath.Join(testRoot, testDir, "out/AVH"))
		assert.Equal(intDir, dirs.IntDir)
		assert.Equal(outDir, dirs.OutDir)
	})

	t.Run("test valid directories as arguments", func(t *testing.T) {
		b.Options.IntDir = "cmdOptionsIntDir"
		b.Options.OutDir = "cmdOptionsOutDir"
		dirs, err := b.getDirs("Hello.Debug+AVH")
		assert.Nil(err)
		intDir, _ := filepath.Abs(filepath.Join(testRoot, testDir, "tmp"))
		outDir, _ := filepath.Abs(b.Options.OutDir)
		assert.Equal(intDir, dirs.IntDir)
		assert.Equal(outDir, dirs.OutDir)
	})

	t.Run("test invalid cbuild-idx", func(t *testing.T) {
		b.InputFile = filepath.Join(testRoot, testDir, "invalid-file.cbuild-idx.yml")
		_, err := b.getDirs("Hello.Debug+AVH")
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

		dirs.OutDir = filepath.Join(testRoot, testDir, "OutDir")
		_ = os.MkdirAll(dirs.OutDir, 0755)
		err := b.clean(dirs, vars)
		assert.Error(err)

		dirs.IntDir = filepath.Join(testRoot, testDir, "IntDir")
		_ = os.MkdirAll(dirs.IntDir, 0755)
		err = b.clean(dirs, vars)
		assert.Error(err)
	})

	t.Run("test clean directories", func(t *testing.T) {
		vars.CmakeBin = testRoot + "/bin/cmake"
		dirs.IntDir = filepath.Join(testRoot, testDir, "tmp")
		dirs.OutDir = filepath.Join(testRoot, testDir, "OutDir")
		_ = os.MkdirAll(dirs.IntDir, 0755)
		_ = os.MkdirAll(dirs.OutDir, 0755)
		err := b.clean(dirs, vars)
		assert.Nil(err)
	})
}

func TestBuild(t *testing.T) {
	assert := assert.New(t)
	configs := inittest.GetTestConfigs(testRoot, testDir)

	b := CbuildIdxBuilder{
		builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: filepath.Join(testRoot, testDir, "Hello.cbuild-idx.yml"),
			Options: builder.Options{
				Contexts: []string{"Hello.Debug+AVH"},
				OutDir:   filepath.Join(testRoot, testDir, "OutDir"),
				Packs:    true,
			},
			InstallConfigs: utils.Configurations{
				BinPath: configs.BinPath,
				BinExtn: configs.BinExtn,
				EtcPath: configs.EtcPath,
			},
		},
	}
	t.Run("test build cbuild-idx", func(t *testing.T) {
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test build cbuild-idx quiet", func(t *testing.T) {
		b.Options.Quiet = true
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test build cbuild-idx debug", func(t *testing.T) {
		b.Options.Debug = true
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test rebuild cbuild-idx", func(t *testing.T) {
		b.Options.Rebuild = true
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test build log", func(t *testing.T) {
		b.Options.LogFile = filepath.Join(testRoot, testDir, "log/test.log")
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
		b.Options.OutDir = filepath.Join(testRoot, testDir, "OutDir")
		b.Options.Debug = true
		b.Options.Generator = "Unix Makefiles"
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test clean cbuild-idx", func(t *testing.T) {
		b.Options.Clean = true
		err := b.Build()
		assert.Nil(err)
	})
}
