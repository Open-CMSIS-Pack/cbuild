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
	"errors"
	"path/filepath"

	"github.com/spf13/cobra"
)

var ListContextsCmd = &cobra.Command{
	Use:   "contexts <csolution.yml>",
	Short: "Print list of contexts in a csolution.yml",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fileExtension := filepath.Ext(args[0])
		if fileExtension != ".yml" {
			return errors.New("invalid file argument")
		}

		configs, err := utils.GetInstallConfigs()
		if err != nil {
			return err
		}

		schameCheck, _ := cmd.Flags().GetBool("schema")
		filter, _ := cmd.Flags().GetString("filter")
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

func init() {
	ListContextsCmd.Flags().StringP("filter", "f", "", "filter results (case sensitive, accepts several expressions)")
}
