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

	log "github.com/sirupsen/logrus"
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
			log.Error("executable path was not found")
			return configs, err
		}
	}
	if binPath != "" {
		binPath, _ = filepath.Abs(binPath)
	}

	configs.BinPath = binPath
	etcPath := filepath.Clean(binPath + "/../etc")
	if _, err := os.Stat(etcPath); os.IsNotExist(err) {
		log.Error("etc directory was not found")
		return configs, err
	}
	if etcPath != "" {
		etcPath, _ = filepath.Abs(etcPath)
	}
	configs.EtcPath = etcPath
	return configs, nil
}
