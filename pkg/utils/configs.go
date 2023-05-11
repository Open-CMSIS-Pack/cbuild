/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
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
	binPath := os.Getenv("CMSIS_BUILD_ROOT")
	if binPath == "" {
		binPath, err = GetExecutablePath()
		if err != nil {
			return Configurations{}, err
		}
	}
	if binPath != "" {
		binPath, _ = filepath.Abs(binPath)
	}

	configs.BinPath = binPath
	etcPath := filepath.Clean(binPath + "/../etc")
	if _, err = os.Stat(etcPath); os.IsNotExist(err) {
		err = errors.New(etcPath + " path was not found")
		return Configurations{}, err
	}
	if etcPath != "" {
		etcPath, _ = filepath.Abs(etcPath)
	}
	configs.EtcPath = etcPath
	return configs, nil
}
