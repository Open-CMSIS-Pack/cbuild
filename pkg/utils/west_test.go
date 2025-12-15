/*
 * Copyright (c) 2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	log "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestWestUtils(t *testing.T) {
	assert := assert.New(t)

	compileCommandsContent := `
[
{
  "directory": "/output/zephyr",
  "command": "gcc -c /src/user/main.c",
  "file": "/src/user/main.c",
  "output": "CMakeFiles/zephyr.dir/src/user/main.o"
},
{
  "directory": "/output/zephyr",
  "command": "gcc -c /src/module/lib.c",
  "file": "/src/module/lib.c",
  "output": "CMakeFiles/zephyr.dir/src/module/lib.o"
},
{
  "directory": "/output/zephyr",
  "command": "gcc -c /src/sub-module/sub-lib.c",
  "file": "/src/module/sub-module/sub-lib.c",
  "output": "CMakeFiles/zephyr.dir/src/module/sub-module/sub-lib.o"
}
]
`
	compileCommandsFile := testRoot + "/" + testDir + "/compile_commands.json"
	_ = os.WriteFile(compileCommandsFile, []byte(compileCommandsContent), 0600)

	zephyrModulesContent := `
"cmsis":"/src/module/hal/cmsis":"${ZEPHYR_CMSIS_CMAKE_DIR}"
"module":"/src/module/":"${ZEPHYR_MODULE_CMAKE_DIR}"
"sub-module":"/src/module/sub-module/":"${ZEPHYR_SUB_MODULE_CMAKE_DIR}"
`
	zephyrModulesFile := testRoot + "/" + testDir + "/zephyr_modules.txt"
	_ = os.WriteFile(zephyrModulesFile, []byte(zephyrModulesContent), 0600)

	cbuildFile := testRoot + "/" + testDir + "/cbuild.yml"
	_ = os.WriteFile(cbuildFile, []byte("build:\n  generated-by: cbuild tests\n"), 0600)

	t.Run("test ParseCompileCommandsFile", func(t *testing.T) {
		result, err := ParseCompileCommandsFile(compileCommandsFile)
		assert.Nil(err)
		assert.Equal("/src/user/main.c", result[0].File)
		assert.Equal("/src/module/lib.c", result[1].File)
		assert.Equal("/src/module/sub-module/sub-lib.c", result[2].File)
	})

	t.Run("test ParseModules", func(t *testing.T) {
		result, err := ParseModules(zephyrModulesFile)
		assert.Nil(err)
		assert.Equal("cmsis", result[0].Name)
		assert.Equal("/src/module/hal/cmsis", result[0].Path)
		assert.Equal("module", result[1].Name)
		assert.Equal("/src/module/", result[1].Path)
		assert.Equal("sub-module", result[2].Name)
		assert.Equal("/src/module/sub-module/", result[2].Path)
	})

	t.Run("test GetModule", func(t *testing.T) {
		commands, err := ParseCompileCommandsFile(compileCommandsFile)
		assert.Nil(err)
		modules, err := ParseModules(zephyrModulesFile)
		assert.Nil(err)

		assert.Equal("", GetModule(commands[0].File, modules))
		assert.Equal("module", GetModule(commands[1].File, modules))
		assert.Equal("sub-module", GetModule(commands[2].File, modules))
	})

	t.Run("test AppendFileToGroupUniquely", func(t *testing.T) {
		fileTree := []Filetree{{Group: "App", Files: []string{"/src/main.c"}}}
		AppendFileToGroupUniquely(&fileTree, "App", "/lib/lib.c")
		AppendFileToGroupUniquely(&fileTree, "hal", "/hal/driver.c")
		// duplicate will be filtered off
		AppendFileToGroupUniquely(&fileTree, "hal", "/hal/driver.c")

		assert.Equal([]Filetree{
			{Group: "App", Files: []string{"/src/main.c", "/lib/lib.c"}},
			{Group: "hal", Files: []string{"/hal/driver.c"}}}, fileTree)
		assert.Equal(2, len(fileTree[0].Files))
		assert.Equal(1, len(fileTree[1].Files))
	})

	t.Run("test AddWestFilesToCbuild", func(t *testing.T) {
		westInfo := WestBuildInfo{
			AppPath: "/src/user/",
			OutDir:  testRoot + "/" + testDir,
			Cbuild:  cbuildFile,
		}

		err := AddWestFilesToCbuild(westInfo)
		assert.Nil(err)

		cbuildContent, _ := os.ReadFile(cbuildFile)

		expected := `build:
  generated-by: cbuild tests
  groups:
    - group: App
      files:
        - file: /src/user/main.c
    - group: Zephyr Modules
      groups:
        - group: module
          files:
            - file: /src/module/lib.c
        - group: sub-module
          files:
            - file: /src/module/sub-module/sub-lib.c
`
		assert.Equal(expected, string(cbuildContent))

		// run again to check idempotency
		err = AddWestFilesToCbuild(westInfo)
		assert.Nil(err)
		cbuildContent, _ = os.ReadFile(cbuildFile)
		assert.Equal(expected, string(cbuildContent))
	})

	t.Run("test AddWestFilesToCbuild with wrong cbuild format", func(t *testing.T) {
		westInfo := WestBuildInfo{
			OutDir: testRoot + "/" + testDir,
			Cbuild: cbuildFile,
		}
		_ = os.WriteFile(cbuildFile, []byte("unknown:\n  generated-by: cbuild tests\n"), 0600)
		err := AddWestFilesToCbuild(westInfo)
		assert.EqualError(err, "invalid cbuild format: '"+westInfo.Cbuild+"'")

		_ = os.WriteFile(cbuildFile, []byte(""), 0600)
		err = AddWestFilesToCbuild(westInfo)
		assert.EqualError(err, "invalid cbuild format: '"+westInfo.Cbuild+"'")
	})

	t.Run("test CheckWestSetup", func(t *testing.T) {
		// west tool is add to the PATH in test initialization
		err := CheckWestSetup()
		assert.Nil(err)

		// save and restore log output
		logger := log.StandardLogger().Out
		defer func() { log.SetOutput(logger) }()
		var logBuffer bytes.Buffer
		log.SetOutput(&logBuffer)

		// incorrect environment variable paths
		os.Setenv("VIRTUAL_ENV", "/invalid/virtual/env")
		os.Setenv("ZEPHYR_BASE", "/invalid/zephyr/base")
		logBuffer.Reset()
		err = CheckWestSetup()
		assert.Nil(err)
		assert.True(strings.Contains(logBuffer.String(), "warning cbuild: VIRTUAL_ENV environment variable specifies non-existent directory"))
		assert.True(strings.Contains(logBuffer.String(), "warning cbuild: ZEPHYR_BASE environment variable specifies non-existent directory"))

		// missing environment variables
		os.Unsetenv("VIRTUAL_ENV")
		os.Unsetenv("ZEPHYR_BASE")
		logBuffer.Reset()
		err = CheckWestSetup()
		assert.Nil(err)
		assert.True(strings.Contains(logBuffer.String(), "warning cbuild: missing VIRTUAL_ENV environment variable"))
		assert.True(strings.Contains(logBuffer.String(), "warning cbuild: missing ZEPHYR_BASE environment variable"))

		// missing west tool
		westBin, err := exec.LookPath("west")
		assert.Nil(err)
		os.Remove(westBin)
		err = CheckWestSetup()
		assert.ErrorContains(err, "exec: \"west\": executable file not found")
	})

	os.Remove(compileCommandsFile)
	os.Remove(zephyrModulesFile)
	os.Remove(cbuildFile)
}
