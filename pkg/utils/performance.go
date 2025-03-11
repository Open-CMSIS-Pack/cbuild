//go:build performance

/*
 * Copyright (c) 2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
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
	Example     string       `json:"Example"`
	OS          string       `json:"OS"`
	Arch        string       `json:"Arch"`
	Performance []PerfResult `json:"performance"`
}

// PerformanceTracker is enabled only in performance mode
type PerformanceTracker struct {
	startTime time.Time
	tool      string
	args      string
	results   []PerfResult
	filePath  string
	mutex     sync.Mutex
}

// set the example name
func SetExample(name string) {
	example = name
}

func GetTrackerInstance(outputPath string) *PerformanceTracker {
	once.Do(func() {
		instance = &PerformanceTracker{
			filePath: outputPath,
		}
	})
	return instance
}

// StartTracking initializes performance tracking
func (pt *PerformanceTracker) StartTracking(tool string, args string) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()
	pt.startTime = time.Now()
	pt.tool = tool
	pt.args = args
}

// StopTracking stops tracking and logs the result
func (pt *PerformanceTracker) StopTracking() {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()
	elapsed := time.Since(pt.startTime).Milliseconds()
	pt.results = append(pt.results, PerfResult{
		Tool:   pt.tool,
		Args:   pt.args,
		TimeMS: elapsed,
	})
}

// SaveResults writes all tracking data to the output file and closes it
func (pt *PerformanceTracker) SaveResults() {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	// Read existing data if the file exists
	var existingEntries []PerformanceEntry
	if _, err := os.Stat(pt.filePath); err == nil {
		fileData, err := ioutil.ReadFile(pt.filePath)
		if err == nil {
			json.Unmarshal(fileData, &existingEntries)
		}
	}

	// Append new entry
	newEntry := PerformanceEntry{
		Example:     example,
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		Performance: pt.results,
	}
	existingEntries = append(existingEntries, newEntry)

	// Write updated data back to the file
	jsonData, err := json.MarshalIndent(existingEntries, "", "  ")
	if err != nil {
		log.Errorf("error marshaling JSON: %v", err)
		return
	}

	err = ioutil.WriteFile(pt.filePath, jsonData, 0644)
	if err != nil {
		log.Errorf("error writing to file: %v", err)
	}
}
