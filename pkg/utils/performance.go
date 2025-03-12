//go:build performance

/*
 * Copyright (c) 2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"encoding/json"
	"os"
	"runtime"
	"sync"
	"time"
)

// Singleton for managing file access
var (
	once     sync.Once
	instance *PerformanceTracker
	example  string
)

type PerfResult struct {
	Tool   string `json:"tool"`
	Args   string `json:"args"`
	TimeMS int64  `json:"time_ms"`
}

type PerformanceEntry struct {
	Example     string       `json:"example"`
	OS          string       `json:"os"`
	Arch        string       `json:"arch"`
	Performance []PerfResult `json:"performance"`
}

type PerformanceTracker struct {
	activeTracking []struct {
		startTime time.Time
		tool      string
		args      string
	}
	results  []PerfResult
	filePath string
	mutex    sync.Mutex
}

// set the example name
func SetExample(name string) {
	example = name
}

func GetTrackerInstance(outputPath string) *PerformanceTracker {
	once.Do(func() {
		instance = &PerformanceTracker{filePath: outputPath}
	})
	if instance.filePath == "" {
		instance.filePath = outputPath
	}
	return instance
}

// StartTracking initializes performance tracking
func (pt *PerformanceTracker) StartTracking(tool string, args string) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()
	pt.activeTracking = append(pt.activeTracking, struct {
		startTime time.Time
		tool      string
		args      string
	}{
		startTime: time.Now(),
		tool:      tool,
		args:      args,
	})
}

// StopTracking stops tracking and logs the result
func (pt *PerformanceTracker) StopTracking() {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()
	if len(pt.activeTracking) == 0 {
		// StopTracking called without a matching StartTracking
		return
	}

	// Pop the last tracking entry
	tracking := pt.activeTracking[len(pt.activeTracking)-1]
	pt.activeTracking = pt.activeTracking[:len(pt.activeTracking)-1]

	elapsed := time.Since(tracking.startTime).Milliseconds()
	pt.results = append(pt.results, PerfResult{
		Tool:   tracking.tool,
		Args:   tracking.args,
		TimeMS: elapsed,
	})
}

// SaveResults writes all tracking data to the output file and closes it
func (pt *PerformanceTracker) SaveResults() error {
	pt.mutex.Lock()
	resultsCopy := append([]PerfResult{}, pt.results...)
	pt.mutex.Unlock()

	// Read existing data if the file exists
	var existingEntries []PerformanceEntry
	if _, err := os.Stat(pt.filePath); err == nil {
		fileData, err := os.ReadFile(pt.filePath)
		if err == nil {
			if err := json.Unmarshal(fileData, &existingEntries); err != nil {
				// Failed to unmarshal JSON file, initializing fresh
				existingEntries = []PerformanceEntry{}
			}
		}
	}

	// Append new entry
	newEntry := PerformanceEntry{
		Example:     example,
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		Performance: resultsCopy,
	}
	existingEntries = append(existingEntries, newEntry)

	// Write updated data back to the file
	jsonData, err := json.MarshalIndent(existingEntries, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(pt.filePath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}
