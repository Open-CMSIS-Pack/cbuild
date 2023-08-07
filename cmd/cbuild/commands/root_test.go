/*
 * Copyright (c) 2022-2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package commands_test

import (
	"cbuild/cmd/cbuild/commands"
	"cbuild/pkg/inittest"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../../test"

func init() {
	inittest.TestInitialization(testRoot)
}

func TestCommands(t *testing.T) {
	assert := assert.New(t)
	cprjFile := testRoot + "/run/minimal.cprj"
	csolutionFile := testRoot + "/run/test.csolution.yml"

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
	logDir := testRoot + "/run/log"
	logFile := logDir + "/test.log"

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

	t.Run("test invalid path to log file", func(t *testing.T) {
		os.RemoveAll(logDir)

		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"--log", logFile, "--version"})
		err := cmd.Execute()
		assert.Nil(err)

		_, err = os.Stat(logFile)
		assert.True(os.IsNotExist(err))
	})

	t.Run("test valid path to log file", func(t *testing.T) {
		_ = os.MkdirAll(logDir, 0755)

		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"--log", logFile, "--version"})
		err := cmd.Execute()
		assert.Nil(err)

		_, err = os.Stat(logFile)
		assert.False(os.IsNotExist(err))
	})
}
