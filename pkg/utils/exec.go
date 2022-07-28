/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type RunnerInterface interface {
	ExecuteCommand(program string, quiet bool, args ...string) error
}

type Runner struct{}

func (r Runner) ExecuteCommand(program string, quiet bool, args ...string) error {
	cmd := exec.Command(program, args...)
	cmd.Stdout = log.StandardLogger().Out
	cmd.Stderr = log.StandardLogger().Out
	return cmd.Run()
}
