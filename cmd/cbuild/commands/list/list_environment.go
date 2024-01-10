/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package list

import (
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder/csolution"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	"github.com/spf13/cobra"
)

var ListEnvironmentCmd = &cobra.Command{
	Use:   "environment",
	Short: "Print list of environment configurations",
	Args:  cobra.MaximumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
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
	},
}

func init() {
	ListEnvironmentCmd.DisableFlagsInUseLine = true
}
