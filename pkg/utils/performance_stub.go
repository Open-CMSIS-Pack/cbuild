//go:build !performance

/*
 * Copyright (c) 2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

// PerformanceTracker is a no-op version used in normal mode
type PerformanceTracker struct{}

func SetExample(name string) {}

// GetTrackerInstance returns nil in normal mode
func GetTrackerInstance(outputPath string) *PerformanceTracker {
	return nil
}

// StartTracking does nothing in normal mode
func (pt *PerformanceTracker) StartTracking(tool string, args string) {}

// StopTracking does nothing in normal mode
func (pt *PerformanceTracker) StopTracking() {}

// SaveResults does nothing in normal mode
func (pt *PerformanceTracker) SaveResults() {}
