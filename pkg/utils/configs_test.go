/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"os"
	"runtime"
	"testing"
	"time"

	cp "github.com/otiai10/copy"

	"github.com/stretchr/testify/assert"
)

func init() {
	// Prepare test data
	_ = os.RemoveAll(testRoot + "/run")
	time.Sleep(time.Second)
	_ = os.MkdirAll(testRoot+"/run/bin", 0755)
	_ = os.MkdirAll(testRoot+"/run/etc", 0755)
	_ = cp.Copy(testRoot+"/data/Test.cbuild-idx.yml", testRoot+"/run/Test.cbuild-idx.yml")
}

func TestGetInstallConfigs(t *testing.T) {
	assert := assert.New(t)
	t.Run("test get install configs with CMSIS_BUILD_ROOT", func(t *testing.T) {
		err := os.Setenv("CMSIS_BUILD_ROOT", testRoot+"/run/bin")
		assert.Nil(err)
		configs, err := GetInstallConfigs()
		assert.Nil(err)
		assert.NotEmpty(configs.BinPath)
		assert.NotEmpty(configs.EtcPath)
		if runtime.GOOS == "windows" {
			assert.NotEmpty(configs.BinExtn)
		} else {
			assert.Empty(configs.BinExtn)
		}
	})

	t.Run("test get install configurations without CMSIS_BUILD_ROOT", func(t *testing.T) {
		err := os.Unsetenv("CMSIS_BUILD_ROOT")
		assert.Nil(err)
		_, err = GetInstallConfigs()
		assert.Error(err)
	})
}
