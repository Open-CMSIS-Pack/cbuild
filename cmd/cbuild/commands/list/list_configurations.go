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
	"os"

	"github.com/spf13/cobra"
)

var ListConfigurationsCmd = &cobra.Command{
	Use:   "configurations <csolution.yml>",
	Short: "Print list of configurations in a csolution.yml",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(args[0]); os.IsNotExist(err) {
			return err
		}

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
		return p.ListConfigurations()
	},
}

func init() {
	ListConfigurationsCmd.Flags().StringP("filter", "f", "", "filter results (case sensitive, accepts several expressions)")
}
