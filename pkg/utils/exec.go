/*
 * Copyright (c) 2022-2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type RunnerInterface interface {
	ExecuteCommand(program string, quiet bool, args ...string) (output string, err error)
}

type Runner struct {
	outBytes []byte
	quite    bool
}

func (r *Runner) Write(bytes []byte) (n int, err error) {
	r.outBytes = append(r.outBytes, bytes...)
	if r.quite {
		return len(bytes), nil
	}
	return log.StandardLogger().Out.Write(bytes)
}

func (r Runner) ExecuteCommand(program string, quiet bool, args ...string) (string, error) {
	r.outBytes = nil
	r.quite = quiet
	cmd := exec.Command(program, args...)
	cmd.Stdout = &r
	cmd.Stderr = log.StandardLogger().Out
	err := cmd.Run()
	return string(r.outBytes), err
}
