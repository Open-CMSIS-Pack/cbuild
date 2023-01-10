/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	"cbuild/cmd/cbuild/commands"
	"os"
	"runtime"
	"testing"
	"time"

	cp "github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../../../test"

func init() {
	// Prepare test data
	_ = os.RemoveAll(testRoot + "/run")
	time.Sleep(2 * time.Second)
	_ = cp.Copy(testRoot+"/data", testRoot+"/run")

	_ = os.MkdirAll(testRoot+"/run/bin", 0755)
	_ = os.MkdirAll(testRoot+"/run/etc", 0755)
	_ = os.MkdirAll(testRoot+"/run/packs", 0755)
	_ = os.MkdirAll(testRoot+"/run/IntDir", 0755)
	_ = os.MkdirAll(testRoot+"/run/OutDir", 0755)

	var binExtension string
	if runtime.GOOS == "windows" {
		binExtension = ".exe"
	}
	cbuildgenBin := testRoot + "/run/bin/cbuildgen" + binExtension
	file, _ := os.Create(cbuildgenBin)
	defer file.Close()
	csolutionBin := testRoot + "/run/bin/csolution" + binExtension
	file, _ = os.Create(csolutionBin)
	defer file.Close()
	cpackgetBin := testRoot + "/run/bin/cpackget" + binExtension
	file, _ = os.Create(cpackgetBin)
	defer file.Close()

	_ = cp.Copy(testRoot+"/run/test.Debug+CM0.cprj", testRoot+"/run/OutDir/test.Debug+CM0.cprj")
}

func TestListCommand(t *testing.T) {
	assert := assert.New(t)
	os.Setenv("CMSIS_BUILD_ROOT", testRoot+"/run/bin")

	t.Run("No arguments", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("invalid flag", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "--invalid"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("multiple arguments", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "test1", "test2"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test help", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"list", "-h"})
		err := cmd.Execute()
		assert.Nil(err)
	})
}
