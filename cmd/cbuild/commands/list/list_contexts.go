/*
 * Copyright (c) 2023-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package list

import (
	"path/filepath"
	"strings"

	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder/csolution"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
	log "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/logger"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	"github.com/spf13/cobra"
)

func listContexts(cmd *cobra.Command, args []string) error {
	var inputFile string
	argCnt := len(args)
	if argCnt == 0 {
		return errutils.New(errutils.ErrRequireArg, "cbuild list contexts --help")
	} else if argCnt == 1 {
		inputFile = args[0]
	} else {
		err := errutils.New(errutils.ErrInvalidCmdLineArg)
		log.Error(err)
		_ = cmd.Help()
		return err
	}

	fileName := filepath.Base(inputFile)
	expectedExtension := ".csolution.yml"
	if !strings.HasSuffix(fileName, expectedExtension) && !strings.HasSuffix(fileName, ".csolution.yaml") {
		return errutils.New(errutils.ErrInvalidFileExtension, fileName, expectedExtension)
	}

	_, err := utils.FileExists(inputFile)
	if err != nil {
		return err
	}

	configs, err := utils.GetInstallConfigs()
	if err != nil {
		return err
	}

	noSchemaChk, _ := cmd.Flags().GetBool("no-schema-check")
	filter, _ := cmd.Flags().GetString("filter")
	p := csolution.CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner: utils.Runner{},
			Options: builder.Options{
				SchemaChk: !noSchemaChk,
				Filter:    filter,
			},
			InputFile:      args[0],
			InstallConfigs: configs,
		},
	}
	return p.ListContexts()
}

var ListContextsCmd = &cobra.Command{
	Use:   "contexts <name>.csolution.yml [options]",
	Short: "Print list of contexts in a csolution.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := listContexts(cmd, args)
		if err != nil {
			log.Error(err)
		}
		return err
	},
}

func init() {
	ListContextsCmd.DisableFlagsInUseLine = true
	ListContextsCmd.Flags().StringP("filter", "f", "", "filter results (case sensitive, accepts several expressions)")
}
