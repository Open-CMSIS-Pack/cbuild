/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package csolution

import (
	builder "cbuild/pkg/builder"
	"cbuild/pkg/utils"
	"errors"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	cp "github.com/otiai10/copy"

	"github.com/stretchr/testify/assert"
)

const testRoot = "../../../test"

type RunnerMock struct{}

func (r RunnerMock) ExecuteCommand(program string, quiet bool, args ...string) (string, error) {
	if strings.Contains(program, "csolution") {
		if args[0] == "list" && args[1] == "contexts" {
			return "test.Debug+CM0\r\ntest.Release+CM0", nil
		} else if args[0] == "list" && args[1] == "packs" {
			return "ARM::test:0.0.1\r\nARM::test2:0.0.2", nil
		} else if args[0] == "convert" {
			return "", nil
		}
	} else if strings.Contains(program, "cbuildgen") {
	} else if strings.Contains(program, "cpackget") {
	} else if strings.Contains(program, "cmake") {
	} else if strings.Contains(program, "ninja") {
	} else if strings.Contains(program, "xmllint") {
	} else {
		return "", errors.New("invalid command")
	}
	return "", nil
}

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

func TestListContexts(t *testing.T) {
	assert := assert.New(t)
	os.Setenv("CMSIS_BUILD_ROOT", testRoot+"/run/bin")
	configs, err := utils.GetInstallConfigs()
	assert.Nil(err)

	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:         RunnerMock{},
			InputFile:      testRoot + "/run/TestSolution/test.csolution.yml",
			InstallConfigs: configs,
		},
	}

	t.Run("test list contexts", func(t *testing.T) {
		err := b.ListContexts()
		assert.Nil(err)
	})

	t.Run("test list contexts with filter", func(t *testing.T) {
		b.Options.Filter = "test"
		err := b.ListContexts()
		assert.Nil(err)
	})

	t.Run("test list contexts with schema check", func(t *testing.T) {
		b.Options.Schema = true
		err := b.ListContexts()
		assert.Nil(err)
	})
}

func TestListToolchians(t *testing.T) {
	assert := assert.New(t)
	os.Setenv("CMSIS_BUILD_ROOT", testRoot+"/run/bin")
	configs, err := utils.GetInstallConfigs()
	assert.Nil(err)
	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:         RunnerMock{},
			InputFile:      testRoot + "/run/TestSolution/test.csolution.yml",
			InstallConfigs: configs,
		},
	}

	t.Run("test list toochains", func(t *testing.T) {
		err := b.ListToolchains()
		assert.Nil(err)
	})

	t.Run("test list toochains with filter", func(t *testing.T) {
		b.Options.Filter = "test"
		err := b.ListToolchains()
		assert.Nil(err)
	})

	t.Run("test list toochains with schema check", func(t *testing.T) {
		b.Options.Schema = true
		err := b.ListToolchains()
		assert.Nil(err)
	})
}

func TestBuild(t *testing.T) {
	assert := assert.New(t)
	os.Setenv("CMSIS_BUILD_ROOT", testRoot+"/run/bin")
	os.Setenv("CMSIS_PACK_ROOT", testRoot+"/run/packs")
	configs, err := utils.GetInstallConfigs()
	assert.Nil(err)

	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: testRoot + "/run/TestSolution/test.csolution.yml",
			Options: builder.Options{
				Context: "test.Debug+CM0",
				IntDir:  testRoot + "/run/IntDir",
				OutDir:  testRoot + "/run/OutDir",
				Packs:   true,
			},
			InstallConfigs: configs,
		},
	}

	t.Run("test build cprj schema check", func(t *testing.T) {
		err := b.Build()
		assert.Nil(err)
	})
}

func TestInstallMissingPacks(t *testing.T) {
	assert := assert.New(t)
	os.Setenv("CMSIS_BUILD_ROOT", testRoot+"/run/bin")
	configs, err := utils.GetInstallConfigs()
	assert.Nil(err)

	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:         RunnerMock{},
			InstallConfigs: configs,
		},
	}

	t.Run("test install missing packs", func(t *testing.T) {
		err = b.installMissingPacks()
		assert.Nil(err)
	})
}
