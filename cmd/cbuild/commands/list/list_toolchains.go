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
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func listToolchains(cmd *cobra.Command, args []string) error {
	var inputFile string
	argCnt := len(args)
	if argCnt == 1 {
		inputFile = args[0]

		fileName := filepath.Base(inputFile)
		expectedExtension := ".csolution.yml"
		if !strings.HasSuffix(fileName, expectedExtension) && !strings.HasSuffix(fileName, ".csolution.yaml") {
			return errutils.New(errutils.ErrInvalidFileExtension, fileName, expectedExtension)
		}

		_, err := utils.FileExists(inputFile)
		if err != nil {
			return err
		}
	} else if argCnt > 1 {
		err := errutils.New(errutils.ErrInvalidCmdLineArg)
		log.Error(err)
		_ = cmd.Help()
		return err
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
}

var ListToolchainsCmd = &cobra.Command{
	Use:   "toolchains [<name>.csolution.yml] [options]",
	Short: "Print list of installed toolchains",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := listToolchains(cmd, args)
		if err != nil {
			log.Error(err)
		}
		return err
	},
}

func init() {
	ListToolchainsCmd.DisableFlagsInUseLine = true
	ListToolchainsCmd.Flags().BoolP("verbose", "v", false, "Enable verbose messages")
}
