/*
 * Copyright (c) 2023-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetInstallConfigs(t *testing.T) {
	assert := assert.New(t)
	t.Run("test get install configurations", func(t *testing.T) {
		_, err := GetInstallConfigs()
		assert.Error(err)
	})
}
