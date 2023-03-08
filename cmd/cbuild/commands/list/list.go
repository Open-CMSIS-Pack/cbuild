/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package list

import (
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list <command> [csolution.yml] [flags]",
	Short: "List information",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	ListToolchainsCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		_ = command.Flags().MarkHidden("schema")
		command.Parent().HelpFunc()(command, strings)
	})
	ListContextsCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		_ = command.Flags().MarkHidden("toolchain")
		command.Parent().HelpFunc()(command, strings)
	})
	ListConfigurationsCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		_ = command.Flags().MarkHidden("toolchain")
		command.Parent().HelpFunc()(command, strings)
	})
	ListCmd.AddCommand(ListConfigurationsCmd, ListContextsCmd, ListToolchainsCmd)
}
