/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
)

type Configurations struct {
	BinPath string
	EtcPath string
	BinExtn string
}

func GetInstallConfigs() (configs Configurations, err error) {
	if runtime.GOOS == "windows" {
		configs.BinExtn = ".exe"
	}
	binPath, err := GetExecutablePath()
	if err != nil {
		return Configurations{}, err
	}
	if binPath != "" {
		binPath, _ = filepath.Abs(binPath)
	}

	configs.BinPath = binPath
	etcPath := filepath.Clean(binPath + "/../etc")
	if _, err = os.Stat(etcPath); os.IsNotExist(err) {
		err = errutils.New(errutils.ErrPathNotFound, etcPath)
		return Configurations{}, err
	}
	if etcPath != "" {
		etcPath, _ = filepath.Abs(etcPath)
	}
	configs.EtcPath = etcPath
	return configs, nil
}
