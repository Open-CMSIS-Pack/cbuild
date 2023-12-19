/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"testing"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/inittest"
	"github.com/stretchr/testify/assert"
)

func init() {
	inittest.TestInitialization(testRoot)
}
func TestGetInstallConfigs(t *testing.T) {
	assert := assert.New(t)
	t.Run("test get install configurations", func(t *testing.T) {
		_, err := GetInstallConfigs()
		assert.Error(err)
	})
}
