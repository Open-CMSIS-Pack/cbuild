/*
 * Copyright (c) 2022-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cproject

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	builder "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/inittest"
	utils "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../../test"
const testDir = "cproject"

func init() {
	inittest.TestInitialization(testRoot, testDir)
}

type RunnerMock struct{}

func (r RunnerMock) ExecuteCommand(program string, quiet bool, args ...string) (string, error) {
	if strings.Contains(program, "cbuildgen") {
		if args[0] == "packlist" {
			packlistFile := filepath.Join(testRoot, testDir, "IntDir/minimal.cpinstall")
			file, _ := os.Create(packlistFile)
			defer file.Close()
		} else if args[0] == "cmake" {
			cmakelistFile := filepath.Join(testRoot, testDir, "IntDir/CMakeLists.txt")
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
		b.InputFile = filepath.Join(testRoot, testDir, "minimal.cprj")
		err := b.checkCprj()
		assert.Nil(err)
	})

	t.Run("test existent file, invalid extension", func(t *testing.T) {
		b.InputFile = filepath.Join(testRoot, testDir, "main.c")
		err := b.checkCprj()
		assert.Error(err)
	})

	t.Run("test invalid file", func(t *testing.T) {
		b.InputFile = filepath.Join(testRoot, testDir, "invalid-file.cprj")
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
		b.InputFile = filepath.Join(testRoot, testDir, "minimal.cprj")
		dirs, err := b.getDirs()
		assert.Nil(err)
		intDir, _ := filepath.Abs(filepath.Join(testRoot, testDir, "IntDir"))
		outDir, _ := filepath.Abs(filepath.Join(testRoot, testDir, "OutDir"))
		assert.Equal(intDir, dirs.IntDir)
		assert.Equal(outDir, dirs.OutDir)
	})

	t.Run("test valid directories in cprj", func(t *testing.T) {
		b.InputFile = filepath.Join(testRoot, testDir, "minimal-dirs.cprj")
		dirs, err := b.getDirs()
		assert.Nil(err)
		intDir, _ := filepath.Abs(filepath.Join(testRoot, testDir, "Intermediate"))
		outDir, _ := filepath.Abs(filepath.Join(testRoot, testDir, "Output"))
		assert.Equal(intDir, dirs.IntDir)
		assert.Equal(outDir, dirs.OutDir)
	})

	t.Run("test valid directories as arguments", func(t *testing.T) {
		b.InputFile = filepath.Join(testRoot, testDir, "minimal.cprj")
		b.Options.IntDir = "cmdOptionsIntDir"
		b.Options.OutDir = "cmdOptionsOutDir"
		dirs, err := b.getDirs()
		assert.Nil(err)
		intDir, _ := filepath.Abs(b.Options.IntDir)
		outDir, _ := filepath.Abs(b.Options.OutDir)
		assert.Equal(intDir, dirs.IntDir)
		assert.Equal(outDir, dirs.OutDir)
	})

	t.Run("test invalid cprj", func(t *testing.T) {
		b.InputFile = filepath.Join(testRoot, testDir, "invalid-file.cprj")
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
	var dirs builder.BuildDirs
	var vars builder.InternalVars

	t.Run("test clean directories, invalid tool", func(t *testing.T) {
		vars.CbuildgenBin = testRoot + "/bin/invalid-tool"

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
		vars.CbuildgenBin = testRoot + "/bin/cbuildgen"
		dirs.IntDir = filepath.Join(testRoot, testDir, "IntDir")
		dirs.OutDir = filepath.Join(testRoot, testDir, "OutDir")
		_ = os.MkdirAll(dirs.IntDir, 0755)
		_ = os.MkdirAll(dirs.OutDir, 0755)
		err := b.clean(dirs, vars)
		assert.Nil(err)
	})

	t.Run("test clean non-existent directories", func(t *testing.T) {
		dirs.IntDir = filepath.Join(testRoot, testDir, "non-existent-intdir")
		dirs.OutDir = filepath.Join(testRoot, testDir, "non-existent-outdir")
		err := b.clean(dirs, vars)
		assert.Nil(err)
	})
}

func TestGetInternalVars(t *testing.T) {
	assert := assert.New(t)
	b := CprjBuilder{
		builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: filepath.Join(testRoot, testDir, "minimal.cprj"),
		},
	}
	t.Run("test get internal vars", func(t *testing.T) {

		_, err := b.GetInternalVars()
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
		jobs := b.GetJobs()
		assert.Equal(jobs, runtime.NumCPU())
	})

	t.Run("test get jobs > 0", func(t *testing.T) {
		b.Options.Jobs = 2
		jobs := b.GetJobs()
		assert.Equal(jobs, 2)
	})

	t.Run("test get jobs < 0", func(t *testing.T) {
		b.Options.Jobs = -1
		jobs := b.GetJobs()
		assert.Equal(jobs, runtime.NumCPU())
	})
}

func TestBuild(t *testing.T) {
	assert := assert.New(t)
	configs := inittest.GetTestConfigs(testRoot, testDir)

	b := CprjBuilder{
		builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: filepath.Join(testRoot, testDir, "minimal.cprj"),
			Options: builder.Options{
				IntDir: filepath.Join(testRoot, testDir, "IntDir"),
				OutDir: filepath.Join(testRoot, testDir, "OutDir"),
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
		b.Options.LockFile = filepath.Join(testRoot, testDir, "lockfile.cprj")
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
		b.Options.IntDir = filepath.Join(testRoot, testDir, "IntDir")
		b.Options.OutDir = filepath.Join(testRoot, testDir, "OutDir")
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
			InputFile: filepath.Join(testRoot, testDir, "minimal.cprj"),
			Options: builder.Options{
				IntDir: filepath.Join(testRoot, testDir, "IntDir"),
				OutDir: filepath.Join(testRoot, testDir, "OutDir"),
			},
		},
	}

	t.Run("test build cprj without packs", func(t *testing.T) {
		b.Options.Packs = false
		err := b.Build()
		assert.Error(err)
	})
}
