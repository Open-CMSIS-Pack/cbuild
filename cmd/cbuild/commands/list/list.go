/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package list

import (
	"errors"

	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list <command> <csolution.yml> [flags]",
	Short: "List project information",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		arg := args[0]
		if !(arg == "contexts" || arg == "toolchains") {
			return errors.New("invalid command")
		}
		return cmd.Help()
	},
}

func init() {
	ListCmd.AddCommand(ListContextsCmd, ListToolchainsCmd)
}
