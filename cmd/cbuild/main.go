/*
 * Copyright (c) 2022-2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"os"

	"github.com/Open-CMSIS-Pack/cbuild/v2/cmd/cbuild/commands"
	log "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/logger"
)

func main() {
	log.SetOutput(os.Stdout)

	commands.Version = version
	commands.CopyrightNotice = copyrightNotice

	cmd := commands.NewRootCmd()
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
