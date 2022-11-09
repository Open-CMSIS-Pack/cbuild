/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package builder

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	cp "github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../test"

type RunnerMock struct{}

func (r RunnerMock) ExecuteCommand(program string, quiet bool, args ...string) error {
	if strings.Contains(program, "cbuildgen") {
		if args[0] == "packlist" {
			packlistFile := testRoot + "/run/IntDir/minimal.cpinstall"
			file, _ := os.Create(packlistFile)
			defer file.Close()
		}
	} else if strings.Contains(program, "cpackget") {
	} else if strings.Contains(program, "cmake") {
	} else if strings.Contains(program, "ninja") {
	} else if strings.Contains(program, "xmllint") {
	} else {
		return errors.New("invalid command")
	}
	return nil
}

func init() {
	// Prepare test data
	_ = os.RemoveAll(testRoot + "/run")
	time.Sleep(time.Second)
	_ = cp.Copy(testRoot+"/data", testRoot+"/run")

	_ = os.MkdirAll(testRoot+"/run/bin", 0755)
	_ = os.MkdirAll(testRoot+"/run/etc", 0755)

	var binExtension string
	if runtime.GOOS == "windows" {
		binExtension = ".exe"
	}
	cbuildgenBin := testRoot + "/run/bin/cbuildgen" + binExtension
	file, _ := os.Create(cbuildgenBin)
	defer file.Close()
	cpackgetBin := testRoot + "/run/bin/cpackget" + binExtension
	file, _ = os.Create(cpackgetBin)
	defer file.Close()
}

func TestConfigLog(t *testing.T) {
	assert := assert.New(t)
	b := Builder{Runner: RunnerMock{}}

	t.Run("test normal verbosity level", func(t *testing.T) {
		b.Options.Quiet = false
		b.Options.Debug = false
		b.configLog()
		assert.Equal(log.InfoLevel, log.GetLevel())
	})

	t.Run("test quiet verbosity level", func(t *testing.T) {
		b.Options.Quiet = true
		b.Options.Debug = false
		b.configLog()
		assert.Equal(log.ErrorLevel, log.GetLevel())
	})

	t.Run("test debug debug level", func(t *testing.T) {
		b.Options.Quiet = false
		b.Options.Debug = true
		b.configLog()
		assert.Equal(log.DebugLevel, log.GetLevel())
	})

	logDir := testRoot + "/run/log"
	b.Options.LogFile = logDir + "/test.log"

	t.Run("test invalid path to log file", func(t *testing.T) {
		os.RemoveAll(logDir)
		b.configLog()
		_, err := os.Stat(b.Options.LogFile)
		assert.True(os.IsNotExist(err))
	})

	t.Run("test valid path to log file", func(t *testing.T) {
		_ = os.MkdirAll(logDir, 0755)
		b.configLog()
		_, err := os.Stat(b.Options.LogFile)
		assert.False(os.IsNotExist(err))
	})
}

func TestCheckCprj(t *testing.T) {
	assert := assert.New(t)
	b := Builder{Runner: RunnerMock{}}

	t.Run("test valid cprj", func(t *testing.T) {
		b.CprjFile = testRoot + "/run/minimal.cprj"
		err := b.checkCprj()
		assert.Nil(err)
	})

	t.Run("test existent file, invalid extension", func(t *testing.T) {
		b.CprjFile = testRoot + "/run/main.c"
		err := b.checkCprj()
		assert.Error(err)
	})

	t.Run("test invalid file", func(t *testing.T) {
		b.CprjFile = testRoot + "/run/invalid-file.cprj"
		err := b.checkCprj()
		assert.Error(err)
	})
}

func TestGetDirs(t *testing.T) {
	assert := assert.New(t)
	b := Builder{Runner: RunnerMock{}}

	t.Run("test default directories", func(t *testing.T) {
		b.CprjFile = testRoot + "/run/minimal.cprj"
		dirs, err := b.getDirs()
		assert.Nil(err)
		intDir, _ := filepath.Abs(testRoot + "/run/IntDir")
		outDir, _ := filepath.Abs(testRoot + "/run/OutDir")
		assert.Equal(intDir, dirs.intDir)
		assert.Equal(outDir, dirs.outDir)
	})

	t.Run("test valid directories in cprj", func(t *testing.T) {
		b.CprjFile = testRoot + "/run/minimal-dirs.cprj"
		dirs, err := b.getDirs()
		assert.Nil(err)
		intDir, _ := filepath.Abs(testRoot + "/run/Intermediate")
		outDir, _ := filepath.Abs(testRoot + "/run/Output")
		assert.Equal(intDir, dirs.intDir)
		assert.Equal(outDir, dirs.outDir)
	})

	t.Run("test valid directories as arguments", func(t *testing.T) {
		b.CprjFile = testRoot + "/run/minimal.cprj"
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
		b.CprjFile = testRoot + "/run/invalid-file.cprj"
		_, err := b.getDirs()
		assert.Error(err)
	})
}

func TestClean(t *testing.T) {
	assert := assert.New(t)
	b := Builder{Runner: RunnerMock{}}
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
	b := Builder{Runner: RunnerMock{}}

	t.Run("test get internal vars, with CMSIS_BUILD_ROOT", func(t *testing.T) {
		b.CprjFile = testRoot + "/run/minimal.cprj"
		os.Setenv("CMSIS_BUILD_ROOT", testRoot+"/run/bin")
		vars, err := b.getInternalVars()
		assert.Nil(err)
		assert.NotEmpty(vars.cprjPath)
		assert.NotEmpty(vars.cprjFilename)
		assert.NotEmpty(vars.binPath)
		assert.NotEmpty(vars.etcPath)
		assert.NotEmpty(vars.cbuildgenBin)
		assert.NotEmpty(vars.cpackgetBin)
	})

	t.Run("test get internal vars, without CMSIS_BUILD_ROOT", func(t *testing.T) {
		b.CprjFile = testRoot + "/run/minimal.cprj"
		os.Unsetenv("CMSIS_BUILD_ROOT")
		_, err := b.getInternalVars()
		assert.Error(err)
	})
}

func TestGetJobs(t *testing.T) {
	assert := assert.New(t)
	b := Builder{Runner: RunnerMock{}}

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
	b := Builder{
		Runner:   RunnerMock{},
		CprjFile: testRoot + "/run/minimal.cprj",
		Options: Options{
			IntDir: testRoot + "/run/IntDir",
			OutDir: testRoot + "/run/OutDir",
			Packs:  true,
		},
	}
	os.Setenv("CMSIS_BUILD_ROOT", testRoot+"/run/bin")

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

	t.Run("test build clean target", func(t *testing.T) {
		b.Options.Target = "clean"
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test build makefile generator", func(t *testing.T) {
		b.Options.IntDir = testRoot + "/run/makefiles/IntDir"
		b.Options.OutDir = testRoot + "/run/makefiles/OutDir"
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
	b := Builder{
		Runner:   RunnerMock{},
		CprjFile: testRoot + "/run/minimal.cprj",
		Options: Options{
			IntDir: testRoot + "/run/IntDir",
			OutDir: testRoot + "/run/OutDir",
		},
	}
	os.Setenv("CMSIS_BUILD_ROOT", testRoot+"/run/bin")

	t.Run("test build cprj without packs", func(t *testing.T) {
		b.Options.Packs = false
		err := b.Build()
		assert.Error(err)
	})
}
