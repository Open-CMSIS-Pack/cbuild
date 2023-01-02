/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package commands

import (
	"cbuild/pkg/builder"
	"cbuild/pkg/builder/csolution"
	"cbuild/pkg/utils"

	"github.com/spf13/cobra"
)

var ListContextsCmd = &cobra.Command{
	Use:   "list-contexts <csolution.yml>",
	Short: "Print list of contexts in a csolution.yml",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		schameCheck, _ := cmd.Flags().GetBool("schema")
		filter, _ := cmd.Flags().GetString("filter")

		configs, err := utils.GetInstallConfigs()
		if err != nil {
			return err
		}

		p := csolution.CSolutionBuilder{
			BuilderParams: builder.BuilderParams{
				Runner: utils.Runner{},
				Options: builder.Options{
					Schema: schameCheck,
					Filter: filter,
				},
				InputFile:      args[0],
				InstallConfigs: configs,
			},
		}
		return p.ListContexts()
	},
}

var ListToolchainsCmd = &cobra.Command{
	Use:   "list-toolchains <csolution.yml>",
	Short: "Print list of installed toolchains",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configs, err := utils.GetInstallConfigs()
		if err != nil {
			return err
		}

		p := csolution.CSolutionBuilder{
			BuilderParams: builder.BuilderParams{
				Runner:         utils.Runner{},
				InputFile:      args[0],
				InstallConfigs: configs,
			},
		}
		return p.ListToolchains()
	},
}

func init() {
	ListContextsCmd.Flags().StringP("filter", "f", "", "filter results (case sensitive, accepts several expressions)")
}
