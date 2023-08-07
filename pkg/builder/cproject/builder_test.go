/*
 * Copyright (c) 2022-2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cproject

import (
	"cbuild/pkg/builder"
	"cbuild/pkg/inittest"
	"cbuild/pkg/utils"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testRoot = "../../../test"

func init() {
	inittest.TestInitialization(testRoot)
}

type RunnerMock struct{}

func (r RunnerMock) ExecuteCommand(program string, quiet bool, args ...string) (string, error) {
	if strings.Contains(program, "cbuildgen") {
		if args[0] == "packlist" {
			packlistFile := testRoot + "/run/IntDir/minimal.cpinstall"
			file, _ := os.Create(packlistFile)
			defer file.Close()
		} else if args[0] == "cmake" {
			cmakelistFile := testRoot + "/run/IntDir/CMakeLists.txt"
			file, _ := os.Create(cmakelistFile)
			defer file.Close()
		}
	} else if strings.Contains(program, "cpackget") {
	} else if strings.Contains(program, "cmake") {
	} else if strings.Contains(program, "ninja") {
	} else if strings.Contains(program, "xmllint") {
	} else {
		return "", errors.New("invalid command")
	}
	return "", nil
}

func TestCheckCprj(t *testing.T) {
	assert := assert.New(t)

	b := CprjBuilder{
		builder.BuilderParams{
			Runner: RunnerMock{},
		},
	}

	t.Run("test valid cprj", func(t *testing.T) {
		b.InputFile = testRoot + "/run/minimal.cprj"
		err := b.checkCprj()
		assert.Nil(err)
	})

	t.Run("test existent file, invalid extension", func(t *testing.T) {
		b.InputFile = testRoot + "/run/main.c"
		err := b.checkCprj()
		assert.Error(err)
	})

	t.Run("test invalid file", func(t *testing.T) {
		b.InputFile = testRoot + "/run/invalid-file.cprj"
		err := b.checkCprj()
		assert.Error(err)
	})
}

func TestGetDirs(t *testing.T) {
	assert := assert.New(t)

	b := CprjBuilder{
		builder.BuilderParams{
			Runner: RunnerMock{},
		},
	}

	t.Run("test default directories", func(t *testing.T) {
		b.InputFile = testRoot + "/run/minimal.cprj"
		dirs, err := b.getDirs()
		assert.Nil(err)
		intDir, _ := filepath.Abs(testRoot + "/run/IntDir")
		outDir, _ := filepath.Abs(testRoot + "/run/OutDir")
		assert.Equal(intDir, dirs.intDir)
		assert.Equal(outDir, dirs.outDir)
	})

	t.Run("test valid directories in cprj", func(t *testing.T) {
		b.InputFile = testRoot + "/run/minimal-dirs.cprj"
		dirs, err := b.getDirs()
		assert.Nil(err)
		intDir, _ := filepath.Abs(testRoot + "/run/Intermediate")
		outDir, _ := filepath.Abs(testRoot + "/run/Output")
		assert.Equal(intDir, dirs.intDir)
		assert.Equal(outDir, dirs.outDir)
	})

	t.Run("test valid directories as arguments", func(t *testing.T) {
		b.InputFile = testRoot + "/run/minimal.cprj"
		b.Options.IntDir = "cmdOptionsIntDir"
		b.Options.OutDir = "cmdOptionsOutDir"
		dirs, err := b.getDirs()
		assert.Nil(err)
		intDir, _ := filepath.Abs(b.Options.IntDir)
		outDir, _ := filepath.Abs(b.Options.OutDir)
		assert.Equal(intDir, dirs.intDir)
		assert.Equal(outDir, dirs.outDir)
	})

	t.Run("test invalid cprj", func(t *testing.T) {
		b.InputFile = testRoot + "/run/invalid-file.cprj"
		_, err := b.getDirs()
		assert.Error(err)
	})
}

func TestClean(t *testing.T) {
	assert := assert.New(t)

	b := CprjBuilder{
		builder.BuilderParams{
			Runner: RunnerMock{},
		},
	}
	var dirs BuildDirs
	var vars InternalVars

	t.Run("test clean directories, invalid tool", func(t *testing.T) {
		vars.cbuildgenBin = testRoot + "/bin/invalid-tool"

		dirs.outDir = testRoot + "/run/OutDir"
		_ = os.MkdirAll(dirs.outDir, 0755)
		err := b.clean(dirs, vars)
		assert.Error(err)

		dirs.intDir = testRoot + "/run/IntDir"
		_ = os.MkdirAll(dirs.intDir, 0755)
		err = b.clean(dirs, vars)
		assert.Error(err)
	})

	t.Run("test clean directories", func(t *testing.T) {
		vars.cbuildgenBin = testRoot + "/bin/cbuildgen"
		dirs.intDir = testRoot + "/run/IntDir"
		dirs.outDir = testRoot + "/run/OutDir"
		_ = os.MkdirAll(dirs.intDir, 0755)
		_ = os.MkdirAll(dirs.outDir, 0755)
		err := b.clean(dirs, vars)
		assert.Nil(err)
	})

	t.Run("test clean non-existent directories", func(t *testing.T) {
		dirs.intDir = testRoot + "/run/non-existent-intdir"
		dirs.outDir = testRoot + "/run/non-existent-outdir"
		err := b.clean(dirs, vars)
		assert.Nil(err)
	})
}

func TestGetInternalVars(t *testing.T) {
	assert := assert.New(t)
	b := CprjBuilder{
		builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: testRoot + "/run/minimal.cprj",
		},
	}
	t.Run("test get internal vars", func(t *testing.T) {

		_, err := b.getInternalVars()
		assert.Error(err)
	})
}

func TestGetJobs(t *testing.T) {
	assert := assert.New(t)
	b := CprjBuilder{
		builder.BuilderParams{
			Runner: RunnerMock{},
		},
	}

	t.Run("test get jobs = 0", func(t *testing.T) {
		b.Options.Jobs = 0
		jobs := b.getJobs()
		assert.Equal(jobs, runtime.NumCPU())
	})

	t.Run("test get jobs > 0", func(t *testing.T) {
		b.Options.Jobs = 2
		jobs := b.getJobs()
		assert.Equal(jobs, 2)
	})

	t.Run("test get jobs < 0", func(t *testing.T) {
		b.Options.Jobs = -1
		jobs := b.getJobs()
		assert.Equal(jobs, runtime.NumCPU())
	})
}

func TestBuild(t *testing.T) {
	assert := assert.New(t)
	configs := inittest.GetTestConfigs(testRoot)

	b := CprjBuilder{
		builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: testRoot + "/run/minimal.cprj",
			Options: builder.Options{
				IntDir: testRoot + "/run/IntDir",
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

	t.Run("test build cprj", func(t *testing.T) {

		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test build cprj schema check", func(t *testing.T) {
		b.Options.Schema = true
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

	t.Run("test build lock file", func(t *testing.T) {
		b.Options.LockFile = testRoot + "/run/lockfile.cprj"
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

	t.Run("test build update rte", func(t *testing.T) {
		b.Options.UpdateRte = true
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test build clean target", func(t *testing.T) {
		b.Options.Target = "clean"
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test build makefile generator", func(t *testing.T) {
		b.Options.IntDir = testRoot + "/run/IntDir"
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

func TestBuildFail(t *testing.T) {
	assert := assert.New(t)
	b := CprjBuilder{
		builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: testRoot + "/run/minimal.cprj",
			Options: builder.Options{
				IntDir: testRoot + "/run/IntDir",
				OutDir: testRoot + "/run/OutDir",
			},
		},
	}

	t.Run("test build cprj without packs", func(t *testing.T) {
		b.Options.Packs = false
		err := b.Build()
		assert.Error(err)
	})
}
