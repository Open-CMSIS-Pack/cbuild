/*
 * Copyright (c) 2022-2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/logger"
)

type RunnerInterface interface {
	ExecuteCommand(program string, quiet bool, args ...string) (output string, err error)
}

type Runner struct {
	outBytes []byte
	quiet    bool
}

func (r *Runner) Write(bytes []byte) (n int, err error) {
	r.outBytes = append(r.outBytes, bytes...)
	if r.quiet {
		return len(bytes), nil
	}
	return log.StandardLogger().Out.Write(bytes)
}

func (r Runner) ExecuteCommand(program string, quiet bool, args ...string) (string, error) {
	// Enable tracking
	tracker := GetTrackerInstance("perf-report.json")
	if tracker != nil {
		tracker.StartTracking(filepath.Base(program), strings.Join(args, " "))
	}

	r.outBytes = nil
	r.quiet = quiet
	cmd := exec.Command(program, args...)
	cmd.Stdout = &r
	cmd.Stderr = log.StandardLogger().Out
	err := cmd.Run()

	// Stop tracking
	if tracker != nil {
		tracker.StopTracking()
	}
	return string(r.outBytes), err
}

// This exclusive function returns the standard output and standard error as strings
func ExecuteCommand(program string, args ...string) (string, string, error) {
	// Enable tracking
	tracker := GetTrackerInstance("")
	if tracker != nil {
		tracker.StartTracking(filepath.Base(program), strings.Join(args, " "))
	}

	cmd := exec.Command(program, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	// Stop tracking
	if tracker != nil {
		tracker.StopTracking()
	}
	return stdout.String(), stderr.String(), err
}
