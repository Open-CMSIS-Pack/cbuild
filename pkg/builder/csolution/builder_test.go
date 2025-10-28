/*
 * Copyright (c) 2023-2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package csolution

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	builder "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/inittest"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../../test"
const testDir = "csolution"

var configs inittest.TestConfigs

func init() {
	inittest.TestInitialization(testRoot, testDir)
	configs = inittest.GetTestConfigs(testRoot, testDir)
}

type RunnerMock struct{}

func (r RunnerMock) ExecuteCommand(program string, quiet bool, args ...string) (string, error) {
	if strings.Contains(program, "csolution") {
		switch args[0] {
		case "list":
			switch args[1] {
			case "contexts":
				return "test.Debug+CM0\r\ntest.Release+CM0", nil
			case "toolchains":
				return "AC5@5.6.7\nAC6@6.18.0\nGCC@11.2.1\nIAR@8.50.6\n", nil
			case "packs":
				return "ARM::test:0.0.1\r\nARM::test2:0.0.2", nil
			case "environment":
				return "CMSIS_PACK_ROOT=C:/Path/Packs\nCMSIS_COMPILER_ROOT=C:/Test/etc\n", nil
			}
		case "convert":
			return "", nil
		}
	} else if strings.Contains(program, "cbuildgen") {
	} else if strings.Contains(program, "cpackget") {
	} else if strings.Contains(program, "cmake") {
	} else if strings.Contains(program, "ninja") {
	} else if strings.Contains(program, "xmllint") {
	} else {
		return "", errutils.New(errutils.ErrInvalidCommand, program)
	}
	return "", nil
}

func TestListContexts(t *testing.T) {
	assert := assert.New(t)
	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: filepath.Join(testRoot, testDir, "TestSolution/test.csolution.yml"),
			InstallConfigs: utils.Configurations{
				BinPath: configs.BinPath,
				BinExtn: configs.BinExtn,
				EtcPath: configs.EtcPath,
			},
		},
	}

	t.Run("test list contexts", func(t *testing.T) {
		contexts, err := b.listContexts(true, false)
		assert.Nil(err)
		assert.Equal(2, len(contexts))
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
		contexts, err := b.listContexts(true, false)
		assert.Nil(err)
		assert.Equal(len(contexts), 2)
		assert.Equal("test.Debug+CM0", contexts[0])
		assert.Equal("test.Release+CM0", contexts[1])
	})
}

func TestListToolchians(t *testing.T) {
	assert := assert.New(t)
	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: filepath.Join(testRoot, testDir, "test.csolution.yml"),
			InstallConfigs: utils.Configurations{
				BinPath: configs.BinPath,
				BinExtn: configs.BinExtn,
				EtcPath: configs.EtcPath,
			},
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
	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner: RunnerMock{},
			InstallConfigs: utils.Configurations{
				BinPath: configs.BinPath,
				BinExtn: configs.BinExtn,
				EtcPath: configs.EtcPath,
			},
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
		b.InstallConfigs = utils.Configurations{
			BinPath: configs.BinPath,
			BinExtn: configs.BinExtn,
			EtcPath: configs.EtcPath,
		}
	})

	t.Run("test list environment", func(t *testing.T) {
		err := b.ListEnvironment()
		assert.Nil(err)
	})

}

func TestBuild(t *testing.T) {
	assert := assert.New(t)
	os.Setenv("CMSIS_PACK_ROOT", filepath.Join(testRoot, testDir, "packs"))
	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: filepath.Join(testRoot, testDir, "Test.csolution.yml"),
			Options: builder.Options{
				OutDir: filepath.Join(testRoot, testDir, "OutDir"),
				Packs:  true,
			},
			InstallConfigs: utils.Configurations{
				BinPath: configs.BinPath,
				BinExtn: configs.BinExtn,
				EtcPath: configs.EtcPath,
			},
		},
	}

	t.Run("test setup", func(t *testing.T) {
		b.Setup = true
		err := b.Build()
		assert.Error(err)
		b.Setup = false
	})

	t.Run("test build csolution without context", func(t *testing.T) {
		err := b.Build()
		assert.Error(err)
	})

	t.Run("test build csolution with context", func(t *testing.T) {
		b.Options.Contexts = []string{"test.Debug+CM0"}
		err := b.Build()
		assert.Error(err)
	})

	t.Run("test build csolution using cbuild2cmake", func(t *testing.T) {
		b.Options.Contexts = []string{}
		b.Options.UseCbuild2CMake = true
		err := b.Build()
		assert.Error(err)
	})

	t.Run("test build csolution with target option", func(t *testing.T) {
		b.Options.Target = "CMakeTarget"
		err := b.Build()
		assert.Error(err)
	})
}

func TestRebuild(t *testing.T) {
	assert := assert.New(t)
	os.Setenv("CMSIS_PACK_ROOT", filepath.Join(testRoot, testDir, "packs"))
	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: filepath.Join(testRoot, testDir, "Test.csolution.yml"),
			Options: builder.Options{
				OutDir:  filepath.Join(testRoot, testDir, "OutDir"),
				Packs:   true,
				Rebuild: true,
			},
			InstallConfigs: utils.Configurations{
				BinPath: configs.BinPath,
				BinExtn: configs.BinExtn,
				EtcPath: configs.EtcPath,
			},
		},
	}

	t.Run("test rebuild csolution without context", func(t *testing.T) {
		err := b.Build()
		assert.Error(err)
	})
}

func TestInstallMissingPacks(t *testing.T) {
	assert := assert.New(t)
	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner: RunnerMock{},
			InstallConfigs: utils.Configurations{
				BinPath: configs.BinPath,
				BinExtn: configs.BinExtn,
				EtcPath: configs.EtcPath,
			},
			Options: builder.Options{
				Packs: true,
			},
		},
	}

	t.Run("test install missing packs", func(t *testing.T) {
		err := b.installMissingPacks()
		assert.Nil(err)
	})

	t.Run("test install missing packs with invalid path", func(t *testing.T) {
		binExtn := b.InstallConfigs.BinExtn
		b.InstallConfigs.BinExtn = "invalid_path"
		err := b.installMissingPacks()
		b.InstallConfigs.BinExtn = binExtn
		assert.Error(err)
	})

	t.Run("test install missing packs with no --pack arg", func(t *testing.T) {
		b.Options.Packs = false
		err := b.installMissingPacks()
		assert.Nil(err)
	})
}

func TestGetCprjFilePath(t *testing.T) {
	assert := assert.New(t)
	testIdxFile := filepath.Join(testRoot, testDir, "Test.cbuild-idx.yml")
	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner: RunnerMock{},
		},
	}

	t.Run("test idx file missing", func(t *testing.T) {
		path, err := b.getCprjFilePath(
			"missingfile.cbuild-idx.yml",
			"HelloWorld_cm0plus.Debug+FRDM-K32L3A6")
		assert.Error(err)
		assert.Equal(path, "")
	})

	t.Run("test get cprj file path with invalid input context", func(t *testing.T) {
		path, err := b.getCprjFilePath(
			testIdxFile,
			"Unknown.Build+Target")
		assert.Error(err)
		assert.Equal(path, "")
	})

	t.Run("test get cprj file path", func(t *testing.T) {
		path, err := b.getCprjFilePath(
			testIdxFile,
			"HelloWorld_cm0plus.Debug+FRDM-K32L3A6")
		assert.Nil(err)
		assert.Equal(path, filepath.Join(testRoot, testDir, "cm0plus", "HelloWorld_cm0plus.Debug+FRDM-K32L3A6.cprj"))
	})
}

func TestGetSelectedContexts(t *testing.T) {
	assert := assert.New(t)
	testSetFile := filepath.Join(testRoot, testDir, "Test.cbuild-set.yml")
	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner: RunnerMock{},
		},
	}

	t.Run("test with missing set file", func(t *testing.T) {
		b.Options.UseContextSet = true
		contexts, err := b.getSelectedContexts("missingfile.cbuild-set.yml")
		assert.Error(err)
		assert.Len(contexts, 0)
	})

	t.Run("test get contexts from set file", func(t *testing.T) {
		expectedContexts := []string{
			"test2.Debug+CM0",
			"test2.Debug+CM3",
			"test1.Debug+CM0",
			"test1.Release+CM0",
		}
		b.Options.UseContextSet = true
		contexts, err := b.getSelectedContexts(testSetFile)
		assert.Nil(err)
		assert.Equal(contexts, expectedContexts)
	})

	t.Run("test get contexts from idx file", func(t *testing.T) {
		expectedContexts := []string{
			"HelloWorld_cm0plus.Debug+FRDM-K32L3A6",
			"HelloWorld_cm0plus.Release+FRDM-K32L3A6",
			"HelloWorld_cm4.Debug+FRDM-K32L3A6",
			"HelloWorld_cm4.Release+FRDM-K32L3A6",
		}
		b.Options.UseContextSet = false
		contexts, err := b.getSelectedContexts(filepath.Join(testRoot, testDir, "Test.cbuild-idx.yml"))
		assert.Nil(err)
		assert.Equal(contexts, expectedContexts)
	})
}

func TestGetIdxFilePath(t *testing.T) {
	assert := assert.New(t)
	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner: RunnerMock{},
		},
	}

	t.Run("test invalid input file", func(t *testing.T) {
		b.InputFile = filepath.Join(testRoot, testDir, "TestSolution/invalid_file.yml")

		path, err := b.getIdxFilePath()
		assert.Error(err)
		assert.Equal(path, "")
	})

	t.Run("test get idx file path", func(t *testing.T) {
		b.InputFile = filepath.Join(testRoot, testDir, "TestSolution/test.csolution.yml")

		path, err := b.getIdxFilePath()
		assert.Nil(err)
		assert.Equal(path, utils.NormalizePath(filepath.Join(testRoot, testDir, "TestSolution/test.cbuild-idx.yml")))
	})
}

func TestFormulateArg(t *testing.T) {
	assert := assert.New(t)
	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: filepath.Join(testRoot, testDir, "Test.csolution.yml"),
		},
	}

	t.Run("test default arg", func(t *testing.T) {
		args := b.formulateArgs([]string{"convert"})
		strArg := utils.NormalizePath(strings.Join(args, " "))
		assert.Equal("convert --solution=../../../test/"+testDir+"/Test.csolution.yml --no-check-schema --no-update-rte", strArg)
	})

	t.Run("test context-set arg", func(t *testing.T) {
		b.Options = builder.Options{
			OutDir:        filepath.Join(testRoot, testDir, "OutDir"),
			Contexts:      []string{"test.Debug+Target", "test.Release+Target"},
			UseContextSet: true,
		}
		args := b.formulateArgs([]string{"convert"})
		strArg := utils.NormalizePath(strings.Join(args, " "))
		assert.Equal("convert --solution=../../../test/"+testDir+"/Test.csolution.yml --no-check-schema --no-update-rte --context=test.Debug+Target --context=test.Release+Target --context-set", strArg)
	})

	t.Run("test --active arg", func(t *testing.T) {
		b.Options = builder.Options{
			TargetSet:    "test",
			UseTargetSet: true,
		}
		args := b.formulateArgs([]string{"convert"})
		strArg := utils.NormalizePath(strings.Join(args, " "))
		assert.Equal("convert --solution=../../../test/"+testDir+"/Test.csolution.yml --no-check-schema --no-update-rte --active=test", strArg)
	})

	t.Run("test --active <empty arg>", func(t *testing.T) {
		b.Options = builder.Options{
			TargetSet:    "",
			UseTargetSet: true,
		}
		args := b.formulateArgs([]string{"convert"})
		strArg := utils.NormalizePath(strings.Join(args, " "))
		assert.Equal("convert --solution=../../../test/"+testDir+"/Test.csolution.yml --no-check-schema --no-update-rte --active=", strArg)
	})
}

func TestGetCbuildSetFilePath(t *testing.T) {
	assert := assert.New(t)

	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner: RunnerMock{},
		},
	}

	t.Run("test get cbuild-set file path", func(t *testing.T) {
		b.InputFile = filepath.Join(testRoot, testDir, "TestSolution/test.csolution.yml")

		path := b.getCbuildSetFilePath()
		assert.Equal(path, utils.NormalizePath(filepath.Join(testRoot, testDir, "TestSolution/test.cbuild-set.yml")))
	})

	t.Run("test get cbuild-set file path with output option", func(t *testing.T) {
		b.InputFile = filepath.Join(testRoot, testDir, "TestSolution/test.csolution.yml")
		b.Options.Output = "OutOfTree"
		path := b.getCbuildSetFilePath()
		assert.Equal(path, utils.NormalizePath("OutOfTree/test.cbuild-set.yml"))
	})
}

func TestHasRebuildNode(t *testing.T) {
	assert := assert.New(t)

	b := CSolutionBuilder{}

	t.Run("test file not found", func(t *testing.T) {
		rebuild, err := b.hasRebuildNode("non_existing_file")
		assert.Error(err)
		assert.False(rebuild)
	})

	t.Run("test no contexts to rebuild", func(t *testing.T) {
		testIdxFile := filepath.Join(testRoot, testDir, "Test.cbuild-idx.yml")

		rebuild, err := b.hasRebuildNode(testIdxFile)
		assert.Nil(err)
		assert.False(rebuild)
	})

	t.Run("test contexts to rebuild found", func(t *testing.T) {
		testIdxFile := filepath.Join(testRoot, testDir, "Rebuild.cbuild-idx.yml")

		rebuild, err := b.hasRebuildNode(testIdxFile)
		assert.Nil(err)
		assert.True(rebuild)
	})
}

func TestIsProjectMoved(t *testing.T) {
	assert := assert.New(t)
	tmpDir := filepath.Join(testRoot, testDir, "tmp")
	_ = os.RemoveAll(tmpDir)

	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: filepath.Join(testRoot, testDir, "Test.csolution.yml"),
		},
	}

	writeTestCMakeCache := func(tmpDirPath, content string) {
		_ = os.RemoveAll(tmpDirPath)
		_ = os.MkdirAll(tmpDirPath, os.ModePerm)
		cmakeCacheFile := filepath.Join(tmpDirPath, "CMakeCache.txt")
		_ = os.WriteFile(cmakeCacheFile, []byte(content), 0600)
	}

	t.Run("test cache file not found", func(t *testing.T) {
		assert.False(b.isProjectMoved())
	})

	t.Run("test project not moved", func(t *testing.T) {
		content := "CMAKE_CACHEFILE_DIR:INTERNAL=" + tmpDir
		writeTestCMakeCache(tmpDir, content)
		defer os.RemoveAll(tmpDir)

		assert.False(b.isProjectMoved())
	})

	t.Run("test project moved", func(t *testing.T) {
		content := "CMAKE_CACHEFILE_DIR:INTERNAL=/home/test/tmp"
		writeTestCMakeCache(tmpDir, content)
		defer os.RemoveAll(tmpDir)

		assert.True(b.isProjectMoved())
	})
}

func TestNeedRebuild(t *testing.T) {
	assert := assert.New(t)
	tmpDir := filepath.Join(testRoot, testDir, "tmp")
	_ = os.RemoveAll(tmpDir)

	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: filepath.Join(testRoot, testDir, "Test.csolution.yml"),
		},
	}

	t.Run("rebuild needed when user specifies -r", func(t *testing.T) {
		b.Options.Rebuild = true
		rebuild, err := b.needRebuild()
		assert.Nil(err)
		assert.False(rebuild)
	})

	t.Run("check rebuild only when --cbuild2cmake", func(t *testing.T) {
		b.Options.Rebuild = false
		b.Options.UseCbuild2CMake = false
		rebuild, err := b.needRebuild()
		assert.Nil(err)
		assert.False(rebuild)
	})

	t.Run("check rebuild needed on new project", func(t *testing.T) {
		b.Options.Rebuild = false
		b.Options.UseCbuild2CMake = true
		rebuild, err := b.needRebuild()
		assert.Nil(err)
		assert.False(rebuild)
	})

	t.Run("check rebuild needed when project moved", func(t *testing.T) {
		b.Options.Rebuild = false
		b.Options.UseCbuild2CMake = true
		content := "CMAKE_CACHEFILE_DIR:INTERNAL=/home/test/tmp"
		_ = os.RemoveAll(tmpDir)
		_ = os.MkdirAll(tmpDir, os.ModePerm)
		cmakeCacheFile := filepath.Join(tmpDir, "CMakeCache.txt")
		_ = os.WriteFile(cmakeCacheFile, []byte(content), 0600)

		rebuild, err := b.needRebuild()
		assert.Nil(err)
		assert.True(rebuild)

		_ = os.RemoveAll(tmpDir)
	})

	t.Run("check rebuild needed when project not moved with no rebuild nodes in idx", func(t *testing.T) {
		b.Options.Rebuild = false
		b.Options.UseCbuild2CMake = true
		content := "CMAKE_CACHEFILE_DIR:INTERNAL=" + tmpDir
		_ = os.RemoveAll(tmpDir)
		_ = os.MkdirAll(tmpDir, os.ModePerm)
		cmakeCacheFile := filepath.Join(tmpDir, "CMakeCache.txt")
		_ = os.WriteFile(cmakeCacheFile, []byte(content), 0600)

		rebuild, err := b.needRebuild()
		assert.Nil(err)
		assert.False(rebuild)

		_ = os.RemoveAll(tmpDir)
	})
}

func TestGetContextsToClean(t *testing.T) {
	assert := assert.New(t)
	b := CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:    RunnerMock{},
			InputFile: filepath.Join(testRoot, testDir, "TestSolution/test.csolution.yml"),
			InstallConfigs: utils.Configurations{
				BinPath: configs.BinPath,
				BinExtn: configs.BinExtn,
				EtcPath: configs.EtcPath,
			},
		},
	}

	t.Run("test get contexts to clean from -c", func(t *testing.T) {
		b.Options.Contexts = []string{"test.Release+CM0"}

		contexts, err := b.getContextsToClean()
		assert.Nil(err)
		assert.Equal(1, len(contexts))
		assert.Equal("test.Release+CM0", contexts[0])
	})

	t.Run("test get contexts to clean from -a", func(t *testing.T) {
		b.InputFile = filepath.Join(testRoot, testDir, "TestSolution/test.csolution.yml")
		b.Options.Contexts = []string{}
		b.Options.TargetSet = "CM0@Custom3"

		contexts, err := b.getContextsToClean()
		assert.Nil(err)
		assert.Equal(1, len(contexts))
		assert.Equal("test.Debug+CM0", contexts[0])
	})

	t.Run("test get contexts to clean with invalid options -c and -a", func(t *testing.T) {
		b.Options.Contexts = []string{"test.Release+CM0"}
		b.Options.TargetSet = "CM0"

		contexts, err := b.getContextsToClean()
		assert.Error(err)
		assert.Equal(0, len(contexts))
	})

	t.Run("test get contexts to clean with invalid options -a and -S", func(t *testing.T) {
		b.Options.Contexts = []string{""}
		b.Options.TargetSet = "CM3"
		b.Options.UseContextSet = true

		contexts, err := b.getContextsToClean()
		assert.Error(err)
		assert.Equal(0, len(contexts))
	})
}
