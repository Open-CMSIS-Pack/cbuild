/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
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
		if args[0] == "list" {
			if args[1] == "contexts" {
				return "test.Debug+CM0\r\ntest.Release+CM0", nil
			} else if args[1] == "toolchains" {
				return "AC5@5.6.7\nAC6@6.18.0\nGCC@11.2.1\nIAR@8.50.6\n", nil
			} else if args[1] == "packs" {
				return "ARM::test:0.0.1\r\nARM::test2:0.0.2", nil
			} else if args[1] == "environment" {
				return "CMSIS_PACK_ROOT=C:/Path/Packs\nCMSIS_COMPILER_ROOT=C:/Test/etc\n", nil
			}
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

func TestListContexts(t *testing.T) {
	assert := assert.New(t)
	os.Setenv("CMSIS_BUILD_ROOT", testRoot+"/run/bin")
	configs, err := utils.GetInstallConfigs()
	assert.Nil(err)

	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:         RunnerMock{},
			InputFile:      testRoot + "/run/test.csolution.yml",
			InstallConfigs: configs,
		},
	}

	t.Run("test list contexts", func(t *testing.T) {
		contexts, err := b.listContexts(true, false)
		assert.Nil(err)
		assert.Equal(len(contexts), 2)
		assert.Equal("test.Debug+CM0", contexts[0])
		assert.Equal("test.Release+CM0", contexts[1])
	})

	t.Run("test list contexts with invalid path", func(t *testing.T) {
		binExtn := b.InstallConfigs.BinExtn
		b.InstallConfigs.BinExtn = "invalid_path"
		_, err := b.listContexts(true, false)
		b.InstallConfigs.BinExtn = binExtn
		assert.Error(err)
	})

	t.Run("test list contexts", func(t *testing.T) {
		err := b.ListContexts()
		assert.Nil(err)
	})

	t.Run("test list contexts with invalid path", func(t *testing.T) {
		binExtn := b.InstallConfigs.BinExtn
		b.InstallConfigs.BinExtn = "invalid_path"
		err := b.ListContexts()
		b.InstallConfigs.BinExtn = binExtn
		assert.Error(err)
	})

	t.Run("test list contexts with filter", func(t *testing.T) {
		b.Options.Filter = "test"
		contexts, err := b.listContexts(true, false)
		assert.Nil(err)
		assert.Equal(len(contexts), 2)
		assert.Equal("test.Debug+CM0", contexts[0])
		assert.Equal("test.Release+CM0", contexts[1])
	})

	t.Run("test list contexts with schema check", func(t *testing.T) {
		b.Options.Schema = true
		contexts, err := b.listContexts(true, false)
		assert.Nil(err)
		assert.Equal(len(contexts), 2)
		assert.Equal("test.Debug+CM0", contexts[0])
		assert.Equal("test.Release+CM0", contexts[1])
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
			InputFile:      testRoot + "/run/test.csolution.yml",
			InstallConfigs: configs,
		},
	}

	t.Run("test list toochains", func(t *testing.T) {
		toolchains, err := b.listToolchains(true)
		assert.Nil(err)
		assert.Equal(len(toolchains), 4)
		assert.Equal("AC5@5.6.7", toolchains[0])
		assert.Equal("AC6@6.18.0", toolchains[1])
		assert.Equal("GCC@11.2.1", toolchains[2])
		assert.Equal("IAR@8.50.6", toolchains[3])
	})

	t.Run("test list toolchains with invalid path", func(t *testing.T) {
		binExtn := b.InstallConfigs.BinExtn
		b.InstallConfigs.BinExtn = "invalid_path"
		_, err := b.listToolchains(true)
		b.InstallConfigs.BinExtn = binExtn
		assert.Error(err)
	})

	t.Run("test list toolchains", func(t *testing.T) {
		err := b.ListToolchains()
		assert.Nil(err)
	})

	t.Run("test list toolchains with invalid path", func(t *testing.T) {
		binExtn := b.InstallConfigs.BinExtn
		b.InstallConfigs.BinExtn = "invalid_path"
		err := b.ListToolchains()
		b.InstallConfigs.BinExtn = binExtn
		assert.Error(err)
	})

	t.Run("test list toochains with filter", func(t *testing.T) {
		b.Options.Filter = "test"
		toolchains, err := b.listToolchains(true)
		assert.Nil(err)
		assert.Equal(len(toolchains), 4)
		assert.Equal("AC5@5.6.7", toolchains[0])
		assert.Equal("AC6@6.18.0", toolchains[1])
		assert.Equal("GCC@11.2.1", toolchains[2])
		assert.Equal("IAR@8.50.6", toolchains[3])
	})

	t.Run("test list toochains with schema check", func(t *testing.T) {
		b.Options.Schema = true
		toolchains, err := b.listToolchains(true)
		assert.Nil(err)
		assert.Equal(len(toolchains), 4)
		assert.Equal("AC5@5.6.7", toolchains[0])
		assert.Equal("AC6@6.18.0", toolchains[1])
		assert.Equal("GCC@11.2.1", toolchains[2])
		assert.Equal("IAR@8.50.6", toolchains[3])
	})
}

func TestListEnvironment(t *testing.T) {
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

	t.Run("test list environment", func(t *testing.T) {
		envConfigs, err := b.listEnvironment(true)
		assert.Nil(err)
		assert.Equal(len(envConfigs), 4)
		assert.Equal("CMSIS_PACK_ROOT=C:/Path/Packs", envConfigs[0])
		assert.Equal("CMSIS_COMPILER_ROOT=C:/Test/etc", envConfigs[1])
		assert.Regexp(`^cmake=([^\s]+)`, envConfigs[2])
		assert.Regexp(`^ninja=([^\s]+)`, envConfigs[3])
	})

	t.Run("test list environment fails to detect", func(t *testing.T) {
		// set empty install config, when cbuild is run standalone (without installation env)
		b.InstallConfigs = utils.Configurations{}
		envConfigs, err := b.listEnvironment(true)
		assert.Error(err)
		assert.Equal(len(envConfigs), 0)
		// restore install config
		b.InstallConfigs = configs
	})

	t.Run("test list environment", func(t *testing.T) {
		err := b.ListEnvironment()
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
			InputFile: testRoot + "/run/test.csolution.yml",
			Options: builder.Options{
				//IntDir: testRoot + "/run/IntDir",
				OutDir: testRoot + "/run/OutDir",
				Packs:  true,
			},
			InstallConfigs: configs,
		},
	}

	t.Run("test build csolution without context", func(t *testing.T) {
		err := b.Build()
		assert.Error(err)
	})

	t.Run("test build csolution with context", func(t *testing.T) {
		b.Options.Context = "test.Debug+CM0"
		err := b.Build()
		assert.Error(err)
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

	t.Run("test install missing packs with invalid path", func(t *testing.T) {
		binExtn := b.InstallConfigs.BinExtn
		b.InstallConfigs.BinExtn = "invalid_path"
		err := b.installMissingPacks()
		b.InstallConfigs.BinExtn = binExtn
		assert.Error(err)
	})
}

// func TestValidateContext(t *testing.T) {
// 	assert := assert.New(t)

// 	allContexts := []string{
// 		"Project1.Build1+Target",
// 		"Project1.Build2+Target",
// 		"Project2.Build1+Target",
// 		"Project2.Build2+Target",
// 		"Project3+Target",
// 	}
// 	testCases := []struct {
// 		Input         string
// 		ExpectError   bool
// 		OutputContext string
// 	}{
// 		// negative test cases
// 		{"", true, ""},
// 		{".+", true, ""},
// 		{".Build1+", true, ""},
// 		{".+Target", true, ""},
// 		{".Build1.Build2+Target", true, ""},
// 		{".Build1+Target+Test", true, ""},
// 		{"+Target", true, ""},
// 		{".Build2", true, ""},
// 		{".Build2+Target", true, ""},
// 		{"+Target.Build1", true, ""},
// 		{"Project", true, ""},
// 		{"Project.Build2", true, ""},
// 		{"Project.Build1+", true, ""},
// 		{"Project+Target.Build1", true, ""},
// 		{"Project1.+Target", true, "Project1+Target"},

// 		// positive test cases
// 		{"Project1.Build2+Target", false, "Project1.Build2+Target"},
// 		{"Project2.Build1+Target", false, "Project2.Build1+Target"},
// 		{"Project3+Target", false, "Project3+Target"},
// 	}
// 	b := CSolutionBuilder{}
// 	for _, test := range testCases {
// 		context, err := b.validateContext(allContexts, test.Input)
// 		if test.ExpectError {
// 			assert.Error(err)
// 		} else {
// 			assert.Nil(err)
// 		}
// 		assert.Equal(context, test.OutputContext)
// 	}
// }
