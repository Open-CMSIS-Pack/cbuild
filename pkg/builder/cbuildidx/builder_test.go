/*
 * Copyright (c) 2024-2026 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cbuildidx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	builder "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/inittest"
	utils "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../../test"
const testDir = "cbuildidx"

func init() {
	inittest.TestInitialization(testRoot, testDir)
}

type RunnerMock struct {
	generateCMakeContexts bool
	generateWestContexts  bool
}

func (r RunnerMock) ExecuteCommand(program string, quiet bool, args ...string) (string, error) {
	if strings.Contains(program, "cbuild2cmake") {
		_ = os.MkdirAll(filepath.Join(testRoot, testDir, "tmp"), 0755)
		cmakelistFile := filepath.Join(testRoot, testDir, "tmp/CMakeLists.txt")
		file, _ := os.Create(cmakelistFile)
		defer file.Close()
		if r.generateCMakeContexts {
			cbuild := "build:\n  context: pending.Release+CM0\n  output-dirs:\n    outdir: ./pending-cmake-out\n"
			_ = os.WriteFile(filepath.Join(testRoot, testDir, "pending.Release+CM0.cbuild.yml"), []byte(cbuild), 0600)
			_ = os.MkdirAll(filepath.Join(testRoot, testDir, "pending-cmake-out"), 0755)
		}
		if r.generateWestContexts {
			cbuild := "build:\n  context: pending.Release+CM0\n  output-dirs:\n    outdir: ./pending-west-out\n  west:\n    app-path: ./pending-west\n"
			_ = os.WriteFile(filepath.Join(testRoot, testDir, "pending.Release+CM0.cbuild.yml"), []byte(cbuild), 0600)
			_ = os.MkdirAll(filepath.Join(testRoot, testDir, "pending-west-out"), 0755)
		}
	} else if strings.Contains(program, "cpackget") {
	} else if strings.Contains(program, "cmake") {
		if len(args) > 0 && args[0] == "--build" {
			buildArgs := strings.Join(args, " ")
			if r.generateCMakeContexts && strings.Contains(buildArgs, "pending.Release+CM0") {
				compileCommands := "[{\"directory\":\".\",\"file\":\"pending/main.c\",\"output\":\"main.c.o\",\"command\":\"cc -c pending/main.c\"}]"
				_ = os.WriteFile(filepath.Join(testRoot, testDir, "pending-cmake-out", "compile_commands.json"), []byte(compileCommands), 0600)
			}
			if r.generateWestContexts && strings.Contains(buildArgs, "pending.Release+CM0") {
				compileCommands := "[{\"directory\":\".\",\"file\":\"pending-west/main.c\",\"output\":\"main.c.o\",\"command\":\"cc -c pending-west/main.c\"}]"
				_ = os.WriteFile(filepath.Join(testRoot, testDir, "pending-west-out", "compile_commands.json"), []byte(compileCommands), 0600)
			}
		}
	} else if strings.Contains(program, "ninja") {
		if args[0] == "--version" {
			return "1.10.2.git.kitware.jobserver-1", nil
		}
	} else if strings.Contains(program, "xmllint") {
	} else {
		return "", errutils.New(errutils.ErrInvalidCommand, program)
	}
	return "", nil
}

func TestGetDirs(t *testing.T) {
	assert := assert.New(t)

	b := CbuildIdxBuilder{
		builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: filepath.Join(testRoot, testDir, "Hello.cbuild-idx.yml"),
		},
	}

	t.Run("test valid directories", func(t *testing.T) {
		dirs, err := b.getDirs("Hello.Debug+AVH")
		assert.Nil(err)
		intDir, _ := filepath.Abs(filepath.Join(testRoot, testDir, "tmp"))
		outDir, _ := filepath.Abs(filepath.Join(testRoot, testDir, "out/AVH"))
		assert.Equal(intDir, dirs.IntDir)
		assert.Equal(outDir, dirs.OutDir)
	})

	t.Run("test valid directories as arguments", func(t *testing.T) {
		b.Options.IntDir = "cmdOptionsIntDir"
		b.Options.OutDir = "cmdOptionsOutDir"
		dirs, err := b.getDirs("Hello.Debug+AVH")
		assert.Nil(err)
		intDir, _ := filepath.Abs(filepath.Join(testRoot, testDir, "tmp"))
		outDir, _ := filepath.Abs(b.Options.OutDir)
		assert.Equal(intDir, dirs.IntDir)
		assert.Equal(outDir, dirs.OutDir)
	})

	t.Run("test invalid cbuild-idx", func(t *testing.T) {
		b.InputFile = filepath.Join(testRoot, testDir, "invalid-file.cbuild-idx.yml")
		_, err := b.getDirs("Hello.Debug+AVH")
		assert.Error(err)
	})
}

func TestBuild(t *testing.T) {
	assert := assert.New(t)
	configs := inittest.GetTestConfigs(testRoot, testDir)

	b := CbuildIdxBuilder{
		builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: filepath.Join(testRoot, testDir, "Hello.cbuild-idx.yml"),
			Options: builder.Options{
				Contexts: []string{"Hello.Debug+AVH"},
				OutDir:   filepath.Join(testRoot, testDir, "OutDir"),
				Packs:    true,
			},
			InstallConfigs: utils.Configurations{
				BinPath: configs.BinPath,
				BinExtn: configs.BinExtn,
				EtcPath: configs.EtcPath,
			},
		},
	}
	t.Run("test build cbuild-idx", func(t *testing.T) {
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test build cbuild-idx quiet", func(t *testing.T) {
		b.Options.Quiet = true
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test build cbuild-idx debug", func(t *testing.T) {
		b.Options.Debug = true
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test rebuild cbuild-idx", func(t *testing.T) {
		b.Options.Rebuild = true
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

	t.Run("test build makefile generator", func(t *testing.T) {
		b.Options.OutDir = filepath.Join(testRoot, testDir, "OutDir")
		b.Options.Debug = true
		b.Options.Generator = "Unix Makefiles"
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test clean cbuild-idx", func(t *testing.T) {
		b.Options.Clean = true
		err := b.Build()
		assert.Nil(err)
	})
}

func TestBuildAllContexts(t *testing.T) {
	assert := assert.New(t)
	configs := inittest.GetTestConfigs(testRoot, testDir)

	b := CbuildIdxBuilder{
		builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: filepath.Join(testRoot, testDir, "Hello.cbuild-idx.yml"),
			Options: builder.Options{
				Contexts: []string{}, // = build all contexts
				OutDir:   filepath.Join(testRoot, testDir, "OutDir"),
				Packs:    true,
			},
			InstallConfigs: utils.Configurations{
				BinPath: configs.BinPath,
				BinExtn: configs.BinExtn,
				EtcPath: configs.EtcPath,
			},
			BuildContext: "Hello.Debug+AVH",
		},
	}
	t.Run("test build cbuild-idx", func(t *testing.T) {
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test build target option", func(t *testing.T) {
		b.Options.Target = "Hello.Debug+AVH"
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test build use context set", func(t *testing.T) {
		b.Options.UseContextSet = true
		err := b.Build()
		assert.Nil(err)
		b.Options.UseContextSet = false
	})

	t.Run("test setup", func(t *testing.T) {
		b.Setup = true
		err := b.Build()
		assert.Nil(err)
	})
}

func TestCompareVersion(t *testing.T) {
	const (
		ERROR   = true
		NOERROR = false
	)

	b := CbuildIdxBuilder{
		builder.BuilderParams{
			Runner: RunnerMock{},
		},
	}

	testCases := []struct {
		version1       string
		version2       string
		expectedOutput int
		expectedError  bool
	}{
		{"", "1.2.3", 0, ERROR},
		{"1.2.3", "", 0, ERROR},
		{"1.2.beta", "1.2.3", 0, ERROR},
		{"1.21.beta", "1.2", 0, ERROR},
		{"foo", "1.2.3", 0, ERROR},
		{"1.7rc2", "1.7", 0, ERROR},
		{"1.7", "1.7rc1", 0, ERROR},
		{"1.0-", "1.0", 0, ERROR},
		{"1.2.3", "1.4.5", -1, NOERROR},
		{"1.2-beta", "1.2-beta", 0, NOERROR},
		{"1.2", "1.1.4", 1, NOERROR},
		{"1.2", "1.2-beta", 1, NOERROR},
		{"1.2+foo", "1.2+beta", 0, NOERROR},
		{"1.2.0", "1.2.0-X-1.2.0+metadata~dist", 1, NOERROR},
	}

	for _, test := range testCases {
		output, err := b.compareVersions(test.version1, test.version2)
		if test.expectedError && err == nil {
			t.Errorf("Expected error, got %v", err)
		}

		if !test.expectedError && err != nil {
			t.Errorf("Expected error %v, got %v", test.expectedError, err)
		}

		if test.expectedOutput != output {
			t.Errorf("Expected output value %v, got %v: for input %v", test.expectedOutput, output, test.version1)
		}
	}
}

func TestGetNinjaVersion(t *testing.T) {
	assert := assert.New(t)

	b := CbuildIdxBuilder{
		builder.BuilderParams{
			Runner: RunnerMock{},
		},
	}

	t.Run("found ninja version", func(t *testing.T) {
		version, err := b.getNinjaVersion()
		assert.Nil(err)
		assert.Equal("1.10.2", version)
	})
}

func TestValidateNinjaVersion(t *testing.T) {
	assert := assert.New(t)

	b := CbuildIdxBuilder{
		builder.BuilderParams{
			Runner: RunnerMock{},
		},
	}

	t.Run("validate installed ninja version with outdated", func(t *testing.T) {
		isGreaterorEqual, err := b.validateNinjaVersion("1.11.1")
		assert.Nil(err)
		assert.False(isGreaterorEqual)
	})

	t.Run("validate installed ninja version is greater", func(t *testing.T) {
		isGreaterorEqual, err := b.validateNinjaVersion("1.10.0")
		assert.Nil(err)
		assert.True(isGreaterorEqual)
	})

	t.Run("validate ninja version with equal version", func(t *testing.T) {
		isGreaterorEqual, err := b.validateNinjaVersion("1.10.2")
		assert.Nil(err)
		assert.True(isGreaterorEqual)
	})

	t.Run("validate with invalid version", func(t *testing.T) {
		output, err := b.validateNinjaVersion("1.10rc1")
		assert.Error(err)
		assert.False(output)
	})
}

func TestImageOnly(t *testing.T) {
	assert := assert.New(t)
	configs := inittest.GetTestConfigs(testRoot, testDir)

	b := CbuildIdxBuilder{
		builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: filepath.Join(testRoot, testDir, "image-only.cbuild-idx.yml"),
			Options: builder.Options{
				Contexts: []string{"+AVH"},
				OutDir:   filepath.Join(testRoot, testDir, "OutDir"),
			},
			InstallConfigs: utils.Configurations{
				BinPath: configs.BinPath,
				BinExtn: configs.BinExtn,
				EtcPath: configs.EtcPath,
			},
		},
	}

	t.Run("validate solution has image-only and executes nodes", func(t *testing.T) {
		imageOnly, executes := b.HasImageOnlyAndExecutes()
		assert.True(imageOnly)
		assert.True(executes)
	})

	t.Run("test build image-only", func(t *testing.T) {
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test setup image-only", func(t *testing.T) {
		b.Setup = true
		err := b.Build()
		assert.Nil(err)
	})
}

func TestWestSupport(t *testing.T) {
	assert := assert.New(t)
	configs := inittest.GetTestConfigs(testRoot, testDir)

	b := CbuildIdxBuilder{
		builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: filepath.Join(testRoot, testDir, "west.cbuild-idx.yml"),
			Options: builder.Options{
				Contexts: []string{},
				OutDir:   filepath.Join(testRoot, testDir, "OutDir"),
			},
			InstallConfigs: utils.Configurations{
				BinPath: configs.BinPath,
				BinExtn: configs.BinExtn,
				EtcPath: configs.EtcPath,
			},
		},
	}

	t.Run("validate west build info", func(t *testing.T) {
		b.BuildContext = "west.Debug+CM0"
		isWest, westInfo := b.GetWestBuildInfo()
		assert.True(isWest)
		assert.Len(westInfo, 1)
		assert.Contains(westInfo[0].CbuildData.Build.OutputDirs.Outdir, "OutDir")
		assert.Contains(westInfo[0].Cbuild, "west.Debug+CM0.cbuild.yml")
	})

	t.Run("test build west", func(t *testing.T) {
		err := b.Build()
		assert.Nil(err)
	})

	t.Run("test setup west", func(t *testing.T) {
		b.Setup = true
		err := b.Build()
		assert.Nil(err)
	})
}

func TestWestSupportMultipleContexts(t *testing.T) {
	assert := assert.New(t)
	configs := inittest.GetTestConfigs(testRoot, testDir)
	indexFile := filepath.Join(testRoot, testDir, "west-multiple.cbuild-idx.yml")
	pendingCbuild := filepath.Join(testRoot, testDir, "pending.Release+CM0.cbuild.yml")
	pendingOutDir := filepath.Join(testRoot, testDir, "pending-west-out")
	index := "build-idx:\n  tmpdir: tmp\n  cbuilds:\n    - cbuild: west.Debug+CM0.cbuild.yml\n      west: true\n      project: west\n      configuration: .Debug+CM0\n    - cbuild: pending.Release+CM0.cbuild.yml\n      west: true\n      project: pending\n      configuration: .Release+CM0\n"
	assert.NoError(os.WriteFile(indexFile, []byte(index), 0600))
	t.Cleanup(func() {
		_ = os.Remove(indexFile)
		_ = os.Remove(pendingCbuild)
		_ = os.RemoveAll(pendingOutDir)
	})

	b := CbuildIdxBuilder{
		builder.BuilderParams{
			Runner:    RunnerMock{generateWestContexts: true},
			InputFile: indexFile,
			Options: builder.Options{
				Contexts: []string{"west.Debug+CM0", "pending.Release+CM0"},
				OutDir:   filepath.Join(testRoot, testDir, "OutDir"),
			},
			InstallConfigs: utils.Configurations{
				BinPath: configs.BinPath,
				BinExtn: configs.BinExtn,
				EtcPath: configs.EtcPath,
			},
		},
	}

	b.BuildContext = "west.Debug+CM0"
	isWest, westInfo := b.GetWestBuildInfo()
	assert.True(isWest)
	assert.Len(westInfo, 1)
	assert.Contains(westInfo[0].Cbuild, "west.Debug+CM0.cbuild.yml")

	westCompileCommands := filepath.Join(testRoot, testDir, "OutDir", "compile_commands.json")
	compileCommands := "[{\"directory\":\".\",\"file\":\"west/main.c\",\"output\":\"main.c.o\",\"command\":\"cc -c west/main.c\"}]"
	assert.NoError(os.WriteFile(westCompileCommands, []byte(compileCommands), 0600))
	t.Cleanup(func() { _ = os.Remove(westCompileCommands) })

	assert.NoError(b.Build())
	b.BuildContext = "pending.Release+CM0"
	assert.NoError(b.Build())
	pendingContent, readErr := os.ReadFile(pendingCbuild)
	assert.NoError(readErr)
	assert.Contains(string(pendingContent), "group: Zephyr Sources")
	assert.Contains(string(pendingContent), "pending-west/main.c")

	isWest, westInfo = b.GetWestBuildInfo()
	assert.True(isWest)
	assert.Len(westInfo, 2)
}

func TestCMakeSupport(t *testing.T) {
	assert := assert.New(t)
	configs := inittest.GetTestConfigs(testRoot, testDir)

	b := CbuildIdxBuilder{
		builder.BuilderParams{
			Runner:       RunnerMock{},
			InputFile:    filepath.Join(testRoot, testDir, "cmake.cbuild-idx.yml"),
			BuildContext: "native.Debug+CM0",
			Options: builder.Options{
				Contexts: []string{},
				OutDir:   filepath.Join(testRoot, testDir, "OutDir"),
			},
			InstallConfigs: utils.Configurations{
				BinPath: configs.BinPath,
				BinExtn: configs.BinExtn,
				EtcPath: configs.EtcPath,
			},
		},
	}

	t.Run("validate cmake build info", func(t *testing.T) {
		isCMake, cmakeInfo := b.GetCMakeBuildInfo()
		assert.True(isCMake)
		assert.Len(cmakeInfo, 1)
		assert.Contains(cmakeInfo[0].OutDir, "cmake-out")
		assert.Contains(cmakeInfo[0].Cbuild, "native.Debug+CM0.cbuild.yml")
	})

	t.Run("test build cmake", func(t *testing.T) {
		err := b.Build()
		assert.Nil(err)
		cbuildContent, readErr := os.ReadFile(filepath.Join(testRoot, testDir, "native.Debug+CM0.cbuild.yml"))
		assert.Nil(readErr)
		assert.Contains(string(cbuildContent), "group: CMake Sources")
		assert.Contains(string(cbuildContent), "native/main.c")
	})

	t.Run("test setup cmake", func(t *testing.T) {
		b.Setup = true
		err := b.Build()
		assert.Nil(err)
	})
}

func TestCMakeSupportMultipleContexts(t *testing.T) {
	assert := assert.New(t)
	configs := inittest.GetTestConfigs(testRoot, testDir)
	pendingCbuild := filepath.Join(testRoot, testDir, "pending.Release+CM0.cbuild.yml")
	pendingOutDir := filepath.Join(testRoot, testDir, "pending-cmake-out")
	t.Cleanup(func() {
		_ = os.Remove(pendingCbuild)
		_ = os.RemoveAll(pendingOutDir)
	})

	b := CbuildIdxBuilder{
		builder.BuilderParams{
			Runner:    RunnerMock{generateCMakeContexts: true},
			InputFile: filepath.Join(testRoot, testDir, "cmake.cbuild-idx.yml"),
			Options: builder.Options{
				Contexts: []string{"native.Debug+CM0", "pending.Release+CM0"},
				OutDir:   filepath.Join(testRoot, testDir, "OutDir"),
			},
			InstallConfigs: utils.Configurations{
				BinPath: configs.BinPath,
				BinExtn: configs.BinExtn,
				EtcPath: configs.EtcPath,
			},
		},
	}

	b.BuildContext = "native.Debug+CM0"
	err := b.Build()
	assert.NoError(err)
	b.BuildContext = "pending.Release+CM0"
	err = b.Build()
	assert.NoError(err)
	nativeContent, readErr := os.ReadFile(filepath.Join(testRoot, testDir, "native.Debug+CM0.cbuild.yml"))
	assert.NoError(readErr)
	assert.Contains(string(nativeContent), "group: CMake Sources")
	assert.Contains(string(nativeContent), "native/main.c")
	pendingContent, readErr := os.ReadFile(pendingCbuild)
	assert.NoError(readErr)
	assert.Contains(string(pendingContent), "group: CMake Sources")
	assert.Contains(string(pendingContent), "pending/main.c")

	b.BuildContext = "pending.Release+CM0"
	isCMake, cmakeInfo := b.GetCMakeBuildInfo()
	assert.True(isCMake)
	assert.Len(cmakeInfo, 2)
}
