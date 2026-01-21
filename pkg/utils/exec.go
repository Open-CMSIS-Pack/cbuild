/*
 * Copyright (c) 2022-2026 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
	log "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/logger"
	"golang.org/x/term"

	"github.com/aymanbagabas/go-pty"
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

var isTerminal = func() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func (r Runner) ExecuteCommand(program string, quiet bool, args ...string) (string, error) {
	// Enable tracking
	tracker := GetTrackerInstance("perf-report.json")
	if tracker != nil {
		tracker.StartTracking(filepath.Base(program), strings.Join(args, " "))
	}

	var err error
	if !quiet && isTerminal() {
		// Use pty to preserve colors and interactive output
		ptmx, ptyErr := pty.New()
		if ptyErr == nil {
			defer ptmx.Close()
			w, h, ptyErr := term.GetSize(int(os.Stdout.Fd()))
			if ptyErr == nil && w > 0 && h > 0 {
				_ = ptmx.Resize(w, h)
			}
			cmd := ptmx.Command(program, args...)
			go func() { _, _ = io.Copy(os.Stdout, ptmx) }()
			err = cmd.Run()
			if err == nil {
				code := cmd.ProcessState.ExitCode()
				if code != 0 {
					err = errutils.New(errutils.ErrChildFailed, code)
				}
			}
		}
	} else {
		// os/exec Command when not running in terminal or in quiet mode
		r.outBytes = nil
		r.quiet = quiet
		cmd := exec.Command(program, args...)
		cmd.Stdout = &r
		cmd.Stderr = log.StandardLogger().Out
		err = cmd.Run()
	}

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
