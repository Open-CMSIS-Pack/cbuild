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

var ListContextsCmd = &cobra.Command{
	Use:   "contexts <name>.csolution.yml [options]",
	Short: "Print list of contexts in a csolution.yml",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configs, err := utils.GetInstallConfigs()
		if err != nil {
			return err
		}

		schemaCheck, _ := cmd.Flags().GetBool("schema")
		filter, _ := cmd.Flags().GetString("filter")
		p := csolution.CSolutionBuilder{
			BuilderParams: builder.BuilderParams{
				Runner: utils.Runner{},
				Options: builder.Options{
					Schema: schemaCheck,
					Filter: filter,
				},
				InputFile:      args[0],
				InstallConfigs: configs,
			},
		}
		return p.ListContexts()
	},
}

func init() {
	ListContextsCmd.DisableFlagsInUseLine = true
	ListContextsCmd.Flags().StringP("filter", "f", "", "filter results (case sensitive, accepts several expressions)")
}
