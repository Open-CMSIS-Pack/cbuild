/*
 * Copyright (c) 2023-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package list

import (
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder/csolution"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
	log "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/logger"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	"github.com/spf13/cobra"
)

func listEnvironment(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return errutils.New(errutils.ErrAcceptNoArgs, "cbuild list environment --help")
	}

	configs, err := utils.GetInstallConfigs()
	if err != nil {
		return err
	}

	p := csolution.CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:         utils.Runner{},
			InstallConfigs: configs,
		},
	}
	return p.ListEnvironment()
}

var ListEnvironmentCmd = &cobra.Command{
	Use:   "environment",
	Short: "Print list of environment configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := listEnvironment(cmd, args)
		if err != nil {
			log.Error(err)
		}
		return err
	},
}

func init() {
	ListEnvironmentCmd.DisableFlagsInUseLine = true
}
