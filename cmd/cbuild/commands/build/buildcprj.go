/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package build

import (
	"errors"
	"path/filepath"

	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder/cproject"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	"github.com/spf13/cobra"
)

func BuildCPRJ(cmd *cobra.Command, args []string) error {
	var inputFile string
	if len(args) == 1 {
		inputFile = args[0]
	} else {
		_ = cmd.Help()
		return errors.New("invalid arguments")
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
		return err
	}

	params := builder.BuilderParams{
		Runner:         utils.Runner{},
		Options:        options,
		InputFile:      inputFile,
		InstallConfigs: configs,
	}

	fileExtension := filepath.Ext(inputFile)
	var b builder.IBuilderInterface
	if fileExtension == ".cprj" {
		b = cproject.CprjBuilder{BuilderParams: params}
	} else {
		return errors.New("invalid file argument")
	}

	return b.Build()
}

var BuildCPRJCmd = &cobra.Command{
	Use:   "buildcprj <name>.cprj [options]",
	Short: "Use a *.CPRJ file as build input",
	Args:  cobra.MinimumNArgs(1),
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
