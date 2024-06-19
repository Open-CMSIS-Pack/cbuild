/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package errutils

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		errorFormat string
		args        []interface{}
		expectedErr string
	}{
		{
			name:        "Basic usage",
			errorFormat: "An error occurred: %s",
			args:        []interface{}{"error details"},
			expectedErr: "An error occurred: error details",
		},
		{
			name:        "No arguments",
			errorFormat: "An error occurred",
			args:        nil,
			expectedErr: "An error occurred",
		},
		{
			name:        "Multiple arguments",
			errorFormat: "Error %s: %s",
			args:        []interface{}{"code", "error message"},
			expectedErr: "Error code: error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.errorFormat, tt.args...)
			if err == nil {
				t.Errorf("Expected error, got nil")
				return
			}
			actualErr := err.Error()
			if actualErr != tt.expectedErr {
				t.Errorf("Expected error message %q, got %q", tt.expectedErr, actualErr)
			}
		})
	}
}
