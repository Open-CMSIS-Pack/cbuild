/*
 * Copyright (c) 2025 Arm Limited. All rights reserved.
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

func listTargetSets(cmd *cobra.Command, args []string) error {
	var inputFile string
	argCnt := len(args)
	switch argCnt {
	case 0:
		return errutils.New(errutils.ErrRequireArg, "cbuild list target-sets --help")
	case 1:
		inputFile = args[0]
	default:
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
	quiet, _ := cmd.Flags().GetBool("quiet")
	verbose, _ := cmd.Flags().GetBool("verbose")

	p := csolution.CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner: utils.Runner{},
			Options: builder.Options{
				SchemaChk: !noSchemaChk,
				Filter:    filter,
				Quiet:     quiet,
				Verbose:   verbose,
			},
			InputFile:      args[0],
			InstallConfigs: configs,
		},
	}
	return p.ListTargetSets()
}

var ListTargetSetsCmd = &cobra.Command{
	Use:   "target-sets <name>.csolution.yml [options]",
	Short: "Print list of target-sets in a <name>.csolution.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := listTargetSets(cmd, args)
		if err != nil {
			log.Error(err)
		}
		return err
	},
}

func init() {
	ListTargetSetsCmd.DisableFlagsInUseLine = true
	ListTargetSetsCmd.Flags().StringP("filter", "f", "", "filter words")
	ListTargetSetsCmd.Flags().BoolP("quiet", "q", false, "Run silently, printing only error messages")
	ListTargetSetsCmd.Flags().BoolP("verbose", "v", false, "Enable verbose messages")
}
