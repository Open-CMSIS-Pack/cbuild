/*
 * Copyright (c) 2023-2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package list

import (
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list <command> [<name>.csolution.yml] [options]",
	Short: "List information about environment, toolchains, and contexts",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	ListCmd.DisableFlagsInUseLine = true
	ListToolchainsCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		_ = command.Flags().MarkHidden("schema")
		command.Parent().HelpFunc()(command, strings)
	})
	ListContextsCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		_ = command.Flags().MarkHidden("schema")
		_ = command.Flags().MarkHidden("toolchain")
		command.Parent().HelpFunc()(command, strings)
	})
	ListEnvironmentCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		_ = command.Flags().MarkHidden("schema")
		_ = command.Flags().MarkHidden("toolchain")
		command.Parent().HelpFunc()(command, strings)
	})
	ListTargetSetsCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		_ = command.Flags().MarkHidden("schema")
		_ = command.Flags().MarkHidden("toolchain")
		command.Parent().HelpFunc()(command, strings)
	})
	ListCmd.AddCommand(ListContextsCmd, ListToolchainsCmd, ListEnvironmentCmd, ListTargetSetsCmd)
}
