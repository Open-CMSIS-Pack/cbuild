/*
 * Copyright (c) 2022-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"os"
	"os/exec"

	"github.com/Open-CMSIS-Pack/cbuild/v2/cmd/cbuild/commands"
	log "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/logger"
)

// csolution ErrorCode: https://github.com/Open-CMSIS-Pack/devtools/blob/main/tools/projmgr/include/ProjMgr.h#L20
const VariableNotDefined = 2
const CompilerNotDefined = 3

func main() {
	log.SetOutput(os.Stdout)

	commands.Version = version
	commands.CopyrightNotice = copyrightNotice

	cmd := commands.NewRootCmd()
	err := cmd.Execute()
	if err != nil {
		errCode := err.(*exec.ExitError).ExitCode()
		if errCode == VariableNotDefined || errCode == CompilerNotDefined {
			// forward csolution error code
			os.Exit(errCode)
		}
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
