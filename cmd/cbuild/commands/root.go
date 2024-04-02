/*
 * Copyright (c) 2022-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package commands

import (
	"errors"
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
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	log "github.com/sirupsen/logrus"
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
	log.SetLevel(log.InfoLevel)
	debug, _ := cmd.Flags().GetBool("debug")
	quiet, _ := cmd.Flags().GetBool("quiet")
	logFile, _ := cmd.Flags().GetString("log")

	if debug {
		log.SetLevel(log.DebugLevel)
	} else if quiet {
		log.SetLevel(log.ErrorLevel)
	}
	if logFile != "" {
		parentLogDir := filepath.Dir(logFile)
		if _, err := os.Stat(parentLogDir); os.IsNotExist(err) {
			if err := os.MkdirAll(parentLogDir, 0755); err != nil {
				return err
			}
		}
		file, err := os.Create(logFile)
		if err != nil {
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
				_ = cmd.Help()
				return errors.New("invalid arguments")
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
			schema, _ := cmd.Flags().GetBool("schema")
			packs, _ := cmd.Flags().GetBool("packs")
			rebuild, _ := cmd.Flags().GetBool("rebuild")
			updateRte, _ := cmd.Flags().GetBool("update-rte")
			toolchain, _ := cmd.Flags().GetString("toolchain")
			useContextSet, _ := cmd.Flags().GetBool("context-set")
			frozenPacks, _ := cmd.Flags().GetBool("frozen-packs")
			useCbuild2CMake, _ := cmd.Flags().GetBool("cbuild2cmake")

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
				Schema:          schema,
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
			} else if fileExtension == ".yml" || fileExtension == ".yaml" {
				b = csolution.CSolutionBuilder{BuilderParams: params}
			} else {
				return errors.New("invalid file argument")
			}

			log.Info("Build Invocation " + Version + CopyrightNotice)
			return b.Build()
		},
	}

	cobra.AddTemplateFunc("replaceString", func(s string) string {
		return strings.Replace(strings.Replace(s, "strings  ", "arg [...]", -1), "string ", "arg    ", -1)
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
	rootCmd.Flags().StringP("load", "l", "", "Set policy for packs loading [latest | all | required]")
	rootCmd.Flags().IntP("jobs", "j", 0, "Number of job slots for parallel execution")
	rootCmd.Flags().StringP("target", "t", "", "Optional CMake target name")
	rootCmd.Flags().StringP("output", "O", "", "Set directory for all output files")
	rootCmd.PersistentFlags().BoolP("schema", "s", false, "Validate project input file(s) against schema")
	rootCmd.PersistentFlags().StringP("log", "", "", "Save output messages in a log file")
	rootCmd.PersistentFlags().StringP("toolchain", "", "", "Input toolchain to be used")
	rootCmd.Flags().BoolP("cbuild2cmake", "", false, "Use build information files with cbuild2cmake interface (experimental)")

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
