/*
 * Copyright (c) 2022-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package commands_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild/v2/cmd/cbuild/commands"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/inittest"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../../test"
const testDir = "command"

func init() {
	inittest.TestInitialization(testRoot, testDir)
}

func TestCommands(t *testing.T) {
	assert := assert.New(t)
	cprjFile := filepath.Join(testRoot, testDir, "minimal.cprj")
	csolutionFile := filepath.Join(testRoot, testDir, "test.csolution.yml")

	t.Run("test version", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"--version"})
		err := cmd.Execute()
		assert.Nil(err)
	})

	t.Run("test help", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"--help"})
		err := cmd.Execute()
		assert.Nil(err)
	})

	t.Run("invalid flag", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"--invalid"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("multiple arguments", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{cprjFile, cprjFile})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test CPRJ build", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{cprjFile})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test CSolution build", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{csolutionFile})
		err := cmd.Execute()
		assert.Error(err)
	})
}

func TestPreLogConfiguration(t *testing.T) {
	assert := assert.New(t)
	logDir := filepath.Join(testRoot, testDir, "log")
	logFile := filepath.Join(logDir, "test.log")

	t.Run("test normal verbosity level", func(t *testing.T) {
		// No quiet, No debug
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"--version"})
		err := cmd.Execute()
		assert.Nil(err)
		assert.Equal(log.InfoLevel, log.GetLevel())
	})

	t.Run("test quiet verbosity level", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"--quiet", "--version"})
		err := cmd.Execute()
		assert.Nil(err)
		assert.Equal(log.ErrorLevel, log.GetLevel())
	})

	t.Run("test debug debug level", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"--debug", "--version"})
		err := cmd.Execute()
		assert.Nil(err)
		assert.Equal(log.DebugLevel, log.GetLevel())
	})

	t.Run("test path generation to log file", func(t *testing.T) {
		os.RemoveAll(logDir)

		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"--log", logFile, "--version"})
		err := cmd.Execute()
		assert.Nil(err)

		_, err = os.Stat(logFile)
		assert.False(os.IsNotExist(err))
	})

	t.Run("test invalid log file path error", func(t *testing.T) {
		os.RemoveAll(logDir)
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			_ = os.MkdirAll(logDir, 0755)
		}
		file := logDir + "/temp"
		_, err := os.Create(file)
		assert.Nil(err)

		invalidLogFilePath := file + "/test/logfile.log"
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"--log", invalidLogFilePath, "--version"})
		err = cmd.Execute()

		// Can't make subdirectory of file.
		assert.Error(err)
		_, err = os.Stat(invalidLogFilePath)
		assert.Error(err)
	})

	t.Run("test log file creation error", func(t *testing.T) {
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			_ = os.MkdirAll(logDir, 0755)
		}
		invalidLogFile := logDir

		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"--log", invalidLogFile, "--version"})
		err := cmd.Execute()
		assert.Error(err)

		// log file should get generated
		fileInfo, err := os.Stat(invalidLogFile)
		assert.Nil(err)
		assert.False(fileInfo.Mode().IsRegular())
	})

	t.Run("test valid path to log file", func(t *testing.T) {
		os.RemoveAll(logDir)

		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"--log", logFile, "--version"})
		err := cmd.Execute()
		assert.Nil(err)

		_, err = os.Stat(logFile)
		assert.False(os.IsNotExist(err))
	})
}
