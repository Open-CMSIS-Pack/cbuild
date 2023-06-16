/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package build_test

import (
	"cbuild/cmd/cbuild/commands"
	"os"
	"runtime"
	"testing"

	cp "github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../../../test"

func init() {
	// Prepare test data
	_ = os.RemoveAll(testRoot + "/run")

	var binExtension string
	if runtime.GOOS == "windows" {
		binExtension = ".exe"
	}
	cbuildgenBin := testRoot + "/run/bin/cbuildgen" + binExtension
	file, _ := os.Create(cbuildgenBin)
	defer file.Close()

	_ = cp.Copy(testRoot+"/run/minimal.cprj", testRoot+"/run/minimal.cprj")
}

func TestBuildCPRJCommand(t *testing.T) {
	assert := assert.New(t)
	os.Setenv("CMSIS_BUILD_ROOT", testRoot+"/run/bin")
	cprjFile := testRoot + "/run/minimal.cprj"

	t.Run("multiple arguments", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"buildcprj", cprjFile, cprjFile})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test CPRJ build", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"buildcprj", cprjFile})
		err := cmd.Execute()
		assert.Error(err)
	})
}

func TestPreLogConfiguration(t *testing.T) {
	assert := assert.New(t)
	os.Setenv("CMSIS_BUILD_ROOT", testRoot+"/run/bin")
	logDir := testRoot + "/run/log"
	logFile := logDir + "/test.log"
	cprjFile := testRoot + "/run/minimal.cprj"

	t.Run("test normal verbosity level", func(t *testing.T) {
		// No quiet, No debug
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"buildcprj", cprjFile, "-C"})
		_ = cmd.Execute()
		assert.Equal(log.InfoLevel, log.GetLevel())
	})

	t.Run("test quiet verbosity level", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"buildcprj", cprjFile, "--quiet", "-C"})
		_ = cmd.Execute()
		assert.Equal(log.ErrorLevel, log.GetLevel())
	})

	t.Run("test debug debug level", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"buildcprj", cprjFile, "--debug", "-C"})
		_ = cmd.Execute()
		assert.Equal(log.DebugLevel, log.GetLevel())
	})

	t.Run("test invalid path to log file", func(t *testing.T) {
		os.RemoveAll(logDir)

		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"buildcprj", cprjFile, "--log", logFile, "-C"})
		_ = cmd.Execute()
		_, err := os.Stat(logFile)
		assert.True(os.IsNotExist(err))
	})

	t.Run("test valid path to log file", func(t *testing.T) {
		_ = os.MkdirAll(logDir, 0755)

		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"buildcprj", cprjFile, "--log", logFile, "-C"})
		_ = cmd.Execute()
		_, err := os.Stat(logFile)
		assert.False(os.IsNotExist(err))
	})
}
