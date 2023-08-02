/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
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
	"time"

	cp "github.com/otiai10/copy"
)

func TestInitialization(testRoot string) {
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
}

type TestConfigs struct {
	BinPath string
	EtcPath string
	BinExtn string
}

func GetTestConfigs(testRoot string) (configs TestConfigs) {
	if runtime.GOOS == "windows" {
		configs.BinExtn = ".exe"
	}
	configs.BinPath, _ = filepath.Abs(testRoot + "/run/bin")
	configs.EtcPath, _ = filepath.Abs(testRoot + "/run/etc")
	return configs
}
