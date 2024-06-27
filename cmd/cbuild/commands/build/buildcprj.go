/*
 * Copyright (c) 2023-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package build

import (
	"path/filepath"

	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder/cproject"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func BuildCPRJ(cmd *cobra.Command, args []string) error {
	var inputFile string
	argCnt := len(args)
	if argCnt == 0 {
		err := errutils.New(errutils.ErrRequireArg, "cbuild buildcprj --help")
		log.Error(err)
		return err
	} else if argCnt == 1 {
		inputFile = args[0]
	} else {
		err := errutils.New(errutils.ErrInvalidCmdLineArg)
		log.Error(err)
		_ = cmd.Help()
		return err
	}

	intDir, _ := cmd.Flags().GetString("intdir")
	outDir, _ := cmd.Flags().GetString("outdir")
	lockFile, _ := cmd.Flags().GetString("update")
	logFile, _ := cmd.Flags().GetString("log")
	generator, _ := cmd.Flags().GetString("generator")
	target, _ := cmd.Flags().GetString("target")
	jobs, _ := cmd.Flags().GetInt("jobs")
	quiet, _ := cmd.Flags().GetBool("quiet")
	debug, _ := cmd.Flags().GetBool("debug")
	verbose, _ := cmd.Flags().GetBool("verbose")
	clean, _ := cmd.Flags().GetBool("clean")
	schema, _ := cmd.Flags().GetBool("schema")
	packs, _ := cmd.Flags().GetBool("packs")
	rebuild, _ := cmd.Flags().GetBool("rebuild")
	updateRte, _ := cmd.Flags().GetBool("update-rte")
	toolchain, _ := cmd.Flags().GetString("toolchain")

	options := builder.Options{
		IntDir:    intDir,
		OutDir:    outDir,
		LockFile:  lockFile,
		LogFile:   logFile,
		Generator: generator,
		Target:    target,
		Jobs:      jobs,
		Quiet:     quiet,
		Debug:     debug,
		Verbose:   verbose,
		Clean:     clean,
		Schema:    schema,
		Packs:     packs,
		Rebuild:   rebuild,
		UpdateRte: updateRte,
		Toolchain: toolchain,
	}

	configs, err := utils.GetInstallConfigs()
	if err != nil {
		log.Error(err)
		return err
	}

	params := builder.BuilderParams{
		Runner:         utils.Runner{},
		Options:        options,
		InputFile:      inputFile,
		InstallConfigs: configs,
	}

	fileExtension := filepath.Ext(inputFile)
	expectedExtension := ".cprj"
	var b builder.IBuilderInterface
	if fileExtension == expectedExtension {
		b = cproject.CprjBuilder{BuilderParams: params}
	} else {
		err := errutils.New(errutils.ErrInvalidFileExtension, fileExtension, expectedExtension)
		log.Error(err)
		return err
	}

	_, err = utils.FileExists(inputFile)
	if err != nil {
		log.Error(err)
		return err
	}

	return b.Build()
}

var BuildCPRJCmd = &cobra.Command{
	Use:    "buildcprj <name>.cprj [options]",
	Short:  "Use a *.CPRJ file as build input",
	Hidden: true, // This makes the command hidden
	RunE: func(cmd *cobra.Command, args []string) error {
		return BuildCPRJ(cmd, args)
	},
}

func init() {
	BuildCPRJCmd.DisableFlagsInUseLine = true
	BuildCPRJCmd.Flags().IntP("jobs", "j", 0, "Number of job slots for parallel execution")
	BuildCPRJCmd.Flags().BoolP("help", "h", false, "Print usage")
	BuildCPRJCmd.Flags().BoolP("quiet", "q", false, "Suppress output messages except build invocations")
	BuildCPRJCmd.Flags().BoolP("debug", "d", false, "Enable debug messages")
	BuildCPRJCmd.Flags().BoolP("verbose", "v", false, "Enable verbose messages from toolchain builds")
	BuildCPRJCmd.Flags().BoolP("clean", "C", false, "Remove intermediate and output directories")
	BuildCPRJCmd.Flags().BoolP("packs", "p", false, "Download missing software packs with cpackget")
	BuildCPRJCmd.Flags().BoolP("rebuild", "r", false, "Remove intermediate and output directories and rebuild")
	BuildCPRJCmd.Flags().BoolP("update-rte", "", false, "Update the RTE directory and files")
	BuildCPRJCmd.Flags().BoolP("schema", "s", false, "Validate project input file(s) against schema")
	BuildCPRJCmd.Flags().StringP("target", "t", "", "Optional CMake target name")
	BuildCPRJCmd.Flags().StringP("log", "", "", "Save output messages in a log file")
	BuildCPRJCmd.Flags().StringP("toolchain", "", "", "Input toolchain to be used")
	BuildCPRJCmd.Flags().StringP("intdir", "i", "", "Set directory for intermediate files")
	BuildCPRJCmd.Flags().StringP("outdir", "o", "", "Set directory for output binary files")
	BuildCPRJCmd.Flags().StringP("update", "u", "", "Generate *.cprj file for reproducing current build")
	BuildCPRJCmd.Flags().StringP("generator", "g", "Ninja", "Select build system generator")
}
