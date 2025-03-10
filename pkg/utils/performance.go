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
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Singleton for managing file access
var (
	once     sync.Once
	instance *PerformanceTracker
)

type PerfResult struct {
	Tool   string `json:"tool"`
	Args   string `json:"args"`
	TimeMS int64  `json:"time_ms"`
}

// PerformanceTracker is enabled only in performance mode
type PerformanceTracker struct {
	startTime time.Time
	tool      string
	args      string
	results   []PerfResult
	file      *os.File
	mutex     sync.Mutex
}

// GetTrackerInstance ensures a singleton instance of PerformanceTracker
func GetTrackerInstance(outputPath string) *PerformanceTracker {
	once.Do(func() {
		file, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Error("error opening file: %v\n", err)
			return
		}
		instance = &PerformanceTracker{
			file: file,
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
	if pt.file == nil {
		return
	}

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
	if pt.file == nil {
		return
	}

	jsonData, _ := json.MarshalIndent(pt.results, "", "  ")
	_, _ = pt.file.Write(jsonData)
	_, _ = pt.file.Write([]byte("\n"))
	pt.file.Close()
}
