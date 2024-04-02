/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package setup

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder/csolution"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	"github.com/spf13/cobra"
)

func SetUpProject(cmd *cobra.Command, args []string) error {
	var inputFile string
	if len(args) == 1 {
		inputFile = args[0]
	} else {
		_ = cmd.Help()
		return errors.New("invalid arguments")
	}

	fileName := filepath.Base(inputFile)
	if !strings.HasSuffix(fileName, ".csolution.yml") {
		return errors.New("invalid file argument")
	}

	logFile, _ := cmd.Flags().GetString("log")
	generator, _ := cmd.Flags().GetString("generator")
	target, _ := cmd.Flags().GetString("target")
	contexts, _ := cmd.Flags().GetStringSlice("context")
	load, _ := cmd.Flags().GetString("load")
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
	useContextSet, _ := cmd.Flags().GetBool("context-set")
	frozenPacks, _ := cmd.Flags().GetBool("frozen-packs")
	useCbuild2CMake, _ := cmd.Flags().GetBool("cbuild2cmake")

	options := builder.Options{
		LogFile:         logFile,
		Generator:       generator,
		Target:          target,
		Jobs:            jobs,
		Quiet:           quiet,
		Debug:           debug,
		Verbose:         verbose,
		Clean:           clean,
		Schema:          schema,
		Packs:           packs,
		Rebuild:         rebuild,
		UpdateRte:       updateRte,
		Contexts:        contexts,
		UseContextSet:   useContextSet,
		Load:            load,
		Toolchain:       toolchain,
		FrozenPacks:     frozenPacks,
		UseCbuild2CMake: useCbuild2CMake,
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
		Setup:          true,
	}

	b := csolution.CSolutionBuilder{
		BuilderParams: params,
	}

	return b.Build()
}

var SetUpCmd = &cobra.Command{
	Use:   "setup <name>.csolution.yml [options]",
	Short: "Generate project data for IDE environment",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return SetUpProject(cmd, args)
	},
}

func init() {
	SetUpCmd.DisableFlagsInUseLine = true
	SetUpCmd.Flags().BoolP("help", "h", false, "Print usage")
	SetUpCmd.Flags().BoolP("quiet", "q", false, "Suppress output messages except build invocations")
	SetUpCmd.Flags().BoolP("debug", "d", false, "Enable debug messages")
	SetUpCmd.Flags().BoolP("verbose", "v", false, "Enable verbose messages from toolchain builds")
	SetUpCmd.Flags().BoolP("clean", "C", false, "Remove intermediate and output directories")
	SetUpCmd.Flags().BoolP("packs", "p", false, "Download missing software packs with cpackget")
	SetUpCmd.Flags().BoolP("rebuild", "r", false, "Remove intermediate and output directories and rebuild")
	SetUpCmd.Flags().BoolP("update-rte", "", false, "Update the RTE directory and files")
	SetUpCmd.Flags().BoolP("context-set", "S", false, "Select the context names from cbuild-set.yml for generating the target application")
	SetUpCmd.Flags().BoolP("frozen-packs", "", false, "Pack list and versions from cbuild-pack.yml are fixed and raises errors if it changes")
	SetUpCmd.Flags().StringP("generator", "g", "Ninja", "Select build system generator")
	SetUpCmd.Flags().StringSliceP("context", "c", []string{}, "Input context names [<project-name>][.<build-type>][+<target-type>]")
	SetUpCmd.Flags().StringP("load", "l", "", "Set policy for packs loading [latest | all | required]")
	SetUpCmd.Flags().IntP("jobs", "j", 0, "Number of job slots for parallel execution")
	SetUpCmd.Flags().StringP("target", "t", "", "Optional CMake target name")
	SetUpCmd.Flags().BoolP("schema", "s", true, "Validate project input file(s) against schema")
	SetUpCmd.Flags().StringP("log", "", "", "Save output messages in a log file")
	SetUpCmd.Flags().StringP("toolchain", "", "", "Input toolchain to be used")
	SetUpCmd.Flags().BoolP("cbuild2cmake", "", false, "Use build information files with cbuild2cmake interface (experimental)")
}
