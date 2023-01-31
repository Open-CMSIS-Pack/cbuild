/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package list

import (
	"cbuild/pkg/builder"
	"cbuild/pkg/builder/csolution"
	"cbuild/pkg/utils"

	"github.com/spf13/cobra"
)

var ListToolchainsCmd = &cobra.Command{
	Use:   "toolchains [csolution.yml]",
	Short: "Print list of installed toolchains",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var inputFile string
		if len(args) == 1 {
			inputFile = args[0]
		}

		configs, err := utils.GetInstallConfigs()
		if err != nil {
			return err
		}

		p := csolution.CSolutionBuilder{
			BuilderParams: builder.BuilderParams{
				Runner:         utils.Runner{},
				InputFile:      inputFile,
				InstallConfigs: configs,
			},
		}
		return p.ListToolchains()
	},
}
