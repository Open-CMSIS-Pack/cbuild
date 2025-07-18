/*
 * Copyright (c) 2022-2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Open-CMSIS-Pack/cbuild/v2/cmd/cbuild/commands/build"
	"github.com/Open-CMSIS-Pack/cbuild/v2/cmd/cbuild/commands/list"
	"github.com/Open-CMSIS-Pack/cbuild/v2/cmd/cbuild/commands/setup"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder/cproject"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder/csolution"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	"github.com/sirupsen/logrus"

	log "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/logger"
	"github.com/spf13/cobra"
)

var Version string

var CopyrightNotice string

func printVersion(file io.Writer) {
	fmt.Fprintf(file, "cbuild version %v%v\n", Version, CopyrightNotice)
}

// UsageTemplate returns usage template for the command.
var usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Options:
{{.LocalFlags.FlagUsages | replaceString | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Options:
{{.InheritedFlags.FlagUsages | replaceString | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

func preConfiguration(cmd *cobra.Command, args []string) error {
	// configure log level
	log.SetLevel(logrus.WarnLevel)
	debug, _ := cmd.Flags().GetBool("debug")
	quiet, _ := cmd.Flags().GetBool("quiet")
	verbose, _ := cmd.Flags().GetBool("verbose")
	logFile, _ := cmd.Flags().GetString("log")

	if debug {
		log.SetLevel(logrus.DebugLevel)
	} else if verbose {
		log.SetLevel(logrus.InfoLevel)
	} else if quiet {
		log.SetLevel(logrus.ErrorLevel)
	}

	if logFile != "" {
		parentLogDir := filepath.Dir(logFile)
		if _, err := os.Stat(parentLogDir); os.IsNotExist(err) {
			if err := os.MkdirAll(parentLogDir, 0755); err != nil {
				log.Error(err)
				return err
			}
		}
		file, err := os.Create(logFile)
		if err != nil {
			log.Error(err)
			return err
		}
		multiWriter := io.MultiWriter(os.Stdout, file)
		log.SetOutput(multiWriter)
	}
	return nil
}

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "cbuild [command] <name>.csolution.yml [options]",
		Short:             "cbuild: Build Invocation " + Version + CopyrightNotice,
		SilenceUsage:      true,
		SilenceErrors:     true,
		PersistentPreRunE: preConfiguration,
		Args:              cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			versionFlag, _ := cmd.Flags().GetBool("version")
			if versionFlag {
				printVersion(cmd.OutOrStdout())
				return nil
			}

			var inputFile string
			if len(args) == 1 {
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
			contexts, _ := cmd.Flags().GetStringSlice("context")
			load, _ := cmd.Flags().GetString("load")
			output, _ := cmd.Flags().GetString("output")
			jobs, _ := cmd.Flags().GetInt("jobs")
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
			targetSet, _ := cmd.Flags().GetString("active")

			// set cbuild2cmake as default tool
			useCbuild2CMake := !useCbuildgen

			// -a option is not compatible with -c or -S
			if targetSet != "" && (len(contexts) > 0 || useContextSet) {
				err := errutils.New(errutils.ErrInvalidTargetSetUsage)
				log.Error(err)
				return err
			}

			if jobs <= 0 {
				err := errutils.New(errutils.ErrInvalidNumJobs)
				log.Error(err)
				return err
			}

			options := builder.Options{
				IntDir:          intDir,
				OutDir:          outDir,
				LockFile:        lockFile,
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
			}

			// get builder for supported input file
			b, err := getBuilder(inputFile, params)
			if err != nil {
				log.Error(err)
				return err
			}

			// check if input file exists
			_, err = utils.FileExists(inputFile)
			if err != nil {
				log.Error(err)
				return err
			}

			log.Info("Build Invocation " + Version + CopyrightNotice)

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

			// Perform the build operation and return its result
			return b.Build()
		},
	}

	cobra.AddTemplateFunc("replaceString", func(s string) string {
		return strings.ReplaceAll(strings.ReplaceAll(s, "strings  ", "arg [...]"), "string ", "arg    ")
	})
	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.DisableFlagsInUseLine = true
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.Flags().BoolP("version", "V", false, "Print version")
	rootCmd.Flags().BoolP("help", "h", false, "Print usage")
	rootCmd.Flags().BoolP("quiet", "q", false, "Suppress output messages except build invocations")
	rootCmd.Flags().BoolP("debug", "d", false, "Enable debug messages")
	rootCmd.Flags().BoolP("verbose", "v", false, "Enable verbose messages from toolchain builds")
	rootCmd.Flags().BoolP("clean", "C", false, "Remove intermediate and output directories")
	rootCmd.Flags().BoolP("packs", "p", false, "Download missing software packs with cpackget")
	rootCmd.Flags().BoolP("rebuild", "r", false, "Remove intermediate and output directories and rebuild")
	rootCmd.Flags().BoolP("update-rte", "", false, "Update the RTE directory and files")
	rootCmd.Flags().BoolP("context-set", "S", false, "Select the context names from cbuild-set.yml for generating the target application")
	rootCmd.Flags().BoolP("frozen-packs", "", false, "Pack list and versions from cbuild-pack.yml are fixed and raises errors if it changes")
	rootCmd.Flags().StringP("generator", "g", "Ninja", "Select build system generator")
	rootCmd.Flags().StringSliceP("context", "c", []string{}, "Input context names [<project-name>][.<build-type>][+<target-type>]")
	rootCmd.Flags().StringP("load", "l", "required", "Set policy for packs loading [latest | all | required]")
	rootCmd.Flags().IntP("jobs", "j", 8, "Number of job slots for parallel execution")
	rootCmd.Flags().StringP("target", "t", "", "Optional CMake target name")
	rootCmd.Flags().StringP("output", "O", "", "Base folder for output files, 'outdir' and 'tmpdir' (default \"Same as '*.csolution.yml'\")")
	rootCmd.PersistentFlags().BoolP("schema", "s", false, "Validate project input file(s) against schema [deprecated]")
	rootCmd.PersistentFlags().BoolP("no-schema-check", "n", false, "Skip schema check")
	rootCmd.PersistentFlags().StringP("log", "", "", "Save output messages in a log file")
	rootCmd.PersistentFlags().StringP("toolchain", "", "", "Input toolchain to be used")
	rootCmd.Flags().BoolP("cbuildgen", "", false, "Generate legacy *.cprj files and use cbuildgen backend")
	rootCmd.Flags().StringP("active", "a", "", "Select active target-set: <target-type>[@<set>]")

	// CPRJ specific hidden flags
	rootCmd.Flags().StringP("intdir", "i", "", "Set directory for intermediate files")
	rootCmd.Flags().StringP("outdir", "o", "", "Set directory for output binary files")
	rootCmd.Flags().StringP("update", "u", "", "Generate *.cprj file for reproducing current build")
	_ = rootCmd.Flags().MarkHidden("intdir")
	_ = rootCmd.Flags().MarkHidden("outdir")
	_ = rootCmd.Flags().MarkHidden("update")

	rootCmd.SetFlagErrorFunc(FlagErrorFunc)
	rootCmd.AddCommand(build.BuildCPRJCmd, list.ListCmd, setup.SetUpCmd)
	return rootCmd
}

func FlagErrorFunc(cmd *cobra.Command, err error) error {
	if err != nil {
		log.Error(err)
		_ = cmd.Help()
	}
	return err
}

func getBuilder(inputFile string, params builder.BuilderParams) (builder.IBuilderInterface, error) {
	fileName := filepath.Base(inputFile)

	switch {
	case strings.HasSuffix(fileName, ".csolution.yml") || strings.HasSuffix(fileName, ".csolution.yaml"):
		return csolution.CSolutionBuilder{BuilderParams: params}, nil
	case strings.HasSuffix(fileName, ".cprj"):
		return cproject.CprjBuilder{BuilderParams: params}, nil
	default:
		return nil, errutils.New(errutils.ErrInvalidFileExtension, fileName, ".csolution.yml or .cprj")
	}
}
