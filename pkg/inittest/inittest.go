/*
 * Copyright (c) 2023-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

// This package is used as a common test setup
// avoiding duplicate setup for all the packages
// under test

package inittest

import (
	"os"
	"path/filepath"
	"runtime"

	cp "github.com/otiai10/copy"
)

func TestInitialization(testRoot string, testDir string) {
	CleanUp(testRoot, testDir)
	testDirPath := testRoot + "/" + testDir
	// Prepare test data
	_ = cp.Copy(testRoot+"/data", testDirPath)
	_ = os.MkdirAll(testDirPath+"/bin", 0755)
	_ = os.MkdirAll(testDirPath+"/etc", 0755)
	_ = os.MkdirAll(testDirPath+"/packs", 0755)
	_ = os.MkdirAll(testDirPath+"/IntDir", 0755)
	_ = os.MkdirAll(testDirPath+"/OutDir", 0755)

	var binExtension string
	if runtime.GOOS == "windows" {
		binExtension = ".exe"
	}
	cbuildgenBin := testDirPath + "/bin/cbuildgen" + binExtension
	file, _ := os.Create(cbuildgenBin)
	defer file.Close()
	csolutionBin := testDirPath + "/bin/csolution" + binExtension
	file, _ = os.Create(csolutionBin)
	defer file.Close()
	cpackgetBin := testDirPath + "/bin/cpackget" + binExtension
	file, _ = os.Create(cpackgetBin)
	defer file.Close()
	cbuild2cmakeBin := testDirPath + "/bin/cbuild2cmake" + binExtension
	file, _ = os.Create(cbuild2cmakeBin)
	defer file.Close()
}

func CleanUp(testRoot string, testDir string) {
	testDirPath := testRoot + "/" + testDir
	_ = os.RemoveAll(testDirPath)
}

type TestConfigs struct {
	BinPath string
	EtcPath string
	BinExtn string
}

func GetTestConfigs(testRoot string, testDir string) (configs TestConfigs) {
	if runtime.GOOS == "windows" {
		configs.BinExtn = ".exe"
	}
	configs.BinPath, _ = filepath.Abs(testRoot + "/" + testDir + "/bin")
	configs.EtcPath, _ = filepath.Abs(testRoot + "/" + testDir + "/etc")
	return configs
}
