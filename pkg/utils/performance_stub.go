//go:build !performance

/*
 * Copyright (c) 2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

// PerformanceTracker is a no-op stub version used in normal mode
// to resolve build errors
type PerformanceTracker struct{}

func SetExample(name string) {}

func GetTrackerInstance(outputPath string) *PerformanceTracker {
	return nil
}

func (pt *PerformanceTracker) StartTracking(tool string, args string) {}

func (pt *PerformanceTracker) StopTracking() {}

func (pt *PerformanceTracker) SaveResults() error {
	return nil
}
