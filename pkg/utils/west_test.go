/*
 * Copyright (c) 2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWestUtils(t *testing.T) {
	assert := assert.New(t)

	compileCommandsContent := `
[
{
  "directory": "/output/zephyr",
  "command": "gcc -c /src/main.c",
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

		assert.Equal("sources", GetModule(commands[0].File, modules))
		assert.Equal("module", GetModule(commands[1].File, modules))
		assert.Equal("sub-module", GetModule(commands[2].File, modules))
	})

	t.Run("test AppendFileToGroup", func(t *testing.T) {
		fileTree := []Filetree{{Group: "sources", Files: []string{"/src/main.c"}}}
		AppendFileToGroup(&fileTree, "sources", "/lib/lib.c")
		AppendFileToGroup(&fileTree, "hal", "/hal/driver.c")

		assert.Equal([]Filetree{
			{Group: "sources", Files: []string{"/src/main.c", "/lib/lib.c"}},
			{Group: "hal", Files: []string{"/hal/driver.c"}}}, fileTree)
	})

	t.Run("test AddWestFilesToCbuild", func(t *testing.T) {
		westInfo := WestBuildInfo{
			OutDir: testRoot + "/" + testDir,
			Cbuild: cbuildFile,
		}

		err := AddWestFilesToCbuild(westInfo)
		assert.Nil(err)

		cbuildContent, _ := os.ReadFile(cbuildFile)
		assert.Equal(
			`build:
  generated-by: cbuild tests
  groups:
    - group: sources
      files:
        - file: /src/user/main.c
    - group: module
      files:
        - file: /src/module/lib.c
    - group: sub-module
      files:
        - file: /src/module/sub-module/sub-lib.c
  constructed-files:
    - file: /src/user/main.c
    - file: /src/module/lib.c
    - file: /src/module/sub-module/sub-lib.c
`, string(cbuildContent))

	})

	os.Remove(compileCommandsFile)
	os.Remove(zephyrModulesFile)
	os.Remove(cbuildFile)
}
