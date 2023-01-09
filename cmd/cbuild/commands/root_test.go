/*
 * Copyright (c) 2022-23 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package commands

import (
	"os"
	"runtime"
	"testing"
	"time"

	cp "github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../../test"

func init() {
	// Prepare test data
	_ = os.RemoveAll(testRoot + "/run")
	time.Sleep(time.Second)
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

func TestCommands(t *testing.T) {
	assert := assert.New(t)
	os.Setenv("CMSIS_BUILD_ROOT", testRoot+"/run/bin")
	cprjFile := testRoot + "/run/minimal.cprj"
	csolutionFile := testRoot + "/run/test.csolution.yml"

	t.Run("test version", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{"--version"})
		err := cmd.Execute()
		assert.Nil(err)
	})

	t.Run("test help", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{"--help"})
		err := cmd.Execute()
		assert.Nil(err)
	})

	t.Run("invalid flag", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{"--invalid"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("multiple arguments", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{cprjFile, cprjFile})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test CPRJ build", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{cprjFile})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test CSolution build", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{csolutionFile})
		err := cmd.Execute()
		assert.Error(err)
	})
}
