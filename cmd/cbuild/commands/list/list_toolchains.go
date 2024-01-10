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

var ListToolchainsCmd = &cobra.Command{
	Use:   "toolchains [<name>.csolution.yml] [options]",
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

		verbose, _ := cmd.Flags().GetBool("verbose")

		p := csolution.CSolutionBuilder{
			BuilderParams: builder.BuilderParams{
				Runner: utils.Runner{},
				Options: builder.Options{
					Verbose: verbose,
				},
				InputFile:      inputFile,
				InstallConfigs: configs,
			},
		}
		return p.ListToolchains()
	},
}

func init() {
	ListToolchainsCmd.DisableFlagsInUseLine = true
	ListToolchainsCmd.Flags().BoolP("verbose", "v", false, "Enable verbose messages")
}
