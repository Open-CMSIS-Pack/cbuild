/*
 * Copyright (c) 2023-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package build_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild/v2/cmd/cbuild/commands"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../../../test"
const testDir = "command"

func TestBuildCPRJCommand(t *testing.T) {
	assert := assert.New(t)
	cprjFile := filepath.Join(testRoot, testDir, "minimal.cprj")

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
	logDir := filepath.Join(testRoot, testDir, "log")
	logFile := filepath.Join(logDir, "test.log")
	cprjFile := filepath.Join(testRoot, testDir, "minimal.cprj")

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

	t.Run("test path generation to log file", func(t *testing.T) {
		os.RemoveAll(logDir)

		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"buildcprj", cprjFile, "--log", logFile, "-C"})
		_ = cmd.Execute()
		_, err := os.Stat(logFile)
		assert.False(os.IsNotExist(err))
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
