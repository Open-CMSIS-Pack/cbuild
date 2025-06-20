/*
 * Copyright (c) 2024-2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package setup

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder/csolution"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
	log "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/logger"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func setUpProject(cmd *cobra.Command, args []string) error {
	var inputFile string
	argCnt := len(args)
	switch argCnt {
	case 0:
		err := errutils.New(errutils.ErrRequireArg, "cbuild setup --help")
		log.Error(err)
		return err
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
		err := errutils.New(errutils.ErrInvalidFileExtension, fileName, expectedExtension)
		log.Error(err)
		return err
	}

	_, err := utils.FileExists(inputFile)
	if err != nil {
		log.Error(err)
		return err
	}

	logFile, _ := cmd.Flags().GetString("log")
	generator, _ := cmd.Flags().GetString("generator")
	target, _ := cmd.Flags().GetString("target")
	contexts, _ := cmd.Flags().GetStringSlice("context")
	load, _ := cmd.Flags().GetString("load")
	jobs, _ := cmd.Flags().GetInt("jobs")
	output, _ := cmd.Flags().GetString("output")
	quiet, _ := cmd.Flags().GetBool("quiet")
	debug, _ := cmd.Flags().GetBool("debug")
	verbose, _ := cmd.Flags().GetBool("verbose")
	clean, _ := cmd.Flags().GetBool("clean")
	noSchemaChk, _ := cmd.Flags().GetBool("no-schema-check")
	packs, _ := cmd.Flags().GetBool("packs")
	rebuild, _ := cmd.Flags().GetBool("rebuild")
	updateRte, _ := cmd.Flags().GetBool("update-rte")
	toolchain, _ := cmd.Flags().GetString("toolchain")
	useContextSet, _ := cmd.Flags().GetBool("context-set")
	frozenPacks, _ := cmd.Flags().GetBool("frozen-packs")
	useCbuildgen, _ := cmd.Flags().GetBool("cbuildgen")
	noDatabase, _ := cmd.Flags().GetBool("no-database")
	targetSet, _ := cmd.Flags().GetString("active")

	useCbuild2CMake := !useCbuildgen

	// Option '-a' and '-S' are mutually exclusive
	if len(targetSet) > 0 && useContextSet {
		err = errutils.New(errutils.ErrInvalidSetUpArgs)
		log.Error(err)
		return err
	}

	// Either '-a' or '-S' must be used
	if len(targetSet) == 0 && !useContextSet {
		err = errutils.New(errutils.ErrMissingRequiredArg)
		log.Error(err)
		return err
	}

	if len(targetSet) > 0 && targetSet[0] == '-' {
		err = errutils.New(errutils.ErrInvalidInputArg, "-a")
		log.Error(err)
		return err
	}

	options := builder.Options{
		LogFile:         logFile,
		Generator:       generator,
		Target:          target,
		Jobs:            jobs,
		Quiet:           quiet,
		Debug:           debug,
		Verbose:         verbose,
		Clean:           clean,
		SchemaChk:       !noSchemaChk,
		Packs:           packs,
		Rebuild:         rebuild,
		UpdateRte:       updateRte,
		Contexts:        contexts,
		UseContextSet:   useContextSet,
		Load:            load,
		Output:          output,
		Toolchain:       toolchain,
		FrozenPacks:     frozenPacks,
		UseCbuild2CMake: useCbuild2CMake,
		NoDatabase:      noDatabase,
		TargetSet:       targetSet,
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
		Setup:          true,
	}

	b := csolution.CSolutionBuilder{
		BuilderParams: params,
	}

	// Check if the user only wants to clean the project
	if rebuild || clean {
		// Perform the clean operation
		err := b.Clean()
		if err != nil {
			log.Error(err)
			return err
		}

		// If it's a clean-only operation, return after cleaning
		if clean {
			return nil
		}
	}

	return b.Build()
}

var SetUpCmd = &cobra.Command{
	Use:   "setup <name>.csolution.yml [options]",
	Short: "Generate project data for IDE environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		perfResultFile, _ := cmd.Flags().GetString("perf-report")
		tracker := utils.GetTrackerInstance(perfResultFile)
		if tracker != nil {
			exampleDir, err := utils.GetParentFolder(args[0])
			if err != nil {
				return err
			}
			utils.SetExample(exampleDir)

			flags := []string{}
			cmd.Flags().Visit(func(f *pflag.Flag) {
				flags = append(flags, fmt.Sprintf("--%s=%s", f.Name, f.Value.String()))
			})

			tracker.StartTracking("cbuild", "setup "+
				strings.Join(args, " ")+" "+strings.Join(flags, " "))
		}

		err := setUpProject(cmd, args)

		if tracker != nil {
			tracker.StopTracking()
			// Save all results
			perfErr := tracker.SaveResults()
			if perfErr != nil {
				err := errutils.New(errutils.ErrPerfResults, perfErr.Error())
				log.Error(err)
			}
		}

		return err
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
	SetUpCmd.Flags().IntP("jobs", "j", 8, "Number of job slots for parallel execution")
	SetUpCmd.Flags().StringP("target", "t", "", "Optional CMake target name")
	SetUpCmd.Flags().StringP("output", "O", "", "Add prefix to 'outdir' and 'tmpdir'")
	SetUpCmd.Flags().BoolP("schema", "s", true, "Validate project input file(s) against schema [deprecated]")
	SetUpCmd.Flags().BoolP("no-schema-check", "n", false, "Skip schema check")
	SetUpCmd.Flags().StringP("log", "", "", "Save output messages in a log file")
	SetUpCmd.Flags().StringP("toolchain", "", "", "Input toolchain to be used")
	SetUpCmd.Flags().BoolP("cbuildgen", "", false, "Generate legacy *.cprj files and use cbuildgen backend")
	SetUpCmd.Flags().BoolP("no-database", "", false, "Skip the generation of compile_commands.json files")
	SetUpCmd.Flags().StringP("active", "a", "", "Select active target-set: <target-type>[@<set>]")

	SetUpCmd.Flags().StringP("perf-report", "", "perf-report.json", "output performance report file")
	_ = SetUpCmd.Flags().MarkHidden("perf-report")
}
