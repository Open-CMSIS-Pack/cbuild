/*
 * Copyright (c) 2022-2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package commands

import (
	"cbuild/cmd/cbuild/commands/list"
	"cbuild/pkg/builder"
	"cbuild/pkg/builder/cproject"
	"cbuild/pkg/builder/csolution"
	"cbuild/pkg/utils"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var version string

const copyrightNotice = " (C) 2022 Arm Ltd. and Contributors"

func printVersion(file io.Writer) {
	fmt.Fprintf(file, "cbuild version %v%v\n", version, copyrightNotice)
}

// UsageTemplate returns usage template for the command.
var usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "cbuild [command] <project.cprj|csolution.yml> [flags]",
		Short:         "cbuild: Build Invocation " + version + copyrightNotice,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(0),
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
			context, _ := cmd.Flags().GetString("context")
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
				Context:   context,
				Load:      load,
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
				b = cproject.CPRJBuilder{BuilderParams: params}
			} else if fileExtension == ".yml" {
				b = csolution.CSolutionBuilder{BuilderParams: params}
			} else {
				return errors.New("invalid file argument")
			}

			log.Info("Build Invocation " + version + copyrightNotice)
			return b.Build()
		},
	}

	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.Flags().BoolP("version", "V", false, "Print version")
	rootCmd.Flags().BoolP("help", "h", false, "Print usage")
	rootCmd.Flags().BoolP("quiet", "q", false, "Suppress output messages except build invocations")
	rootCmd.Flags().BoolP("debug", "d", false, "Enable debug messages")
	rootCmd.Flags().BoolP("verbose", "v", false, "Enable verbose messages from toolchain builds")
	rootCmd.Flags().BoolP("clean", "c", false, "Remove intermediate and output directories")
	rootCmd.PersistentFlags().BoolP("schema", "s", false, "Validate project input file(s) against schema")
	rootCmd.Flags().BoolP("packs", "p", false, "Download missing software packs with cpackget")
	rootCmd.Flags().BoolP("rebuild", "r", false, "Remove intermediate and output directories and rebuild")
	rootCmd.Flags().BoolP("update-rte", "", false, "Update the RTE directory and files")
	rootCmd.Flags().StringP("intdir", "i", "", "Set directory for intermediate files")
	rootCmd.Flags().StringP("outdir", "o", "", "Set directory for output files")
	rootCmd.Flags().StringP("update", "u", "", "Generate *.cprj file for reproducing current build")
	rootCmd.Flags().StringP("log", "l", "", "Save output messages in a log file")
	rootCmd.Flags().StringP("generator", "g", "Ninja", "Select build system generator")
	rootCmd.Flags().StringP("context", "", "", "Input context name e.g. project.buildType+targetType")
	rootCmd.Flags().StringP("load", "", "", "Set policy for packs loading [latest|all|required]")
	rootCmd.Flags().IntP("jobs", "j", 0, "Number of job slots for parallel execution")
	rootCmd.Flags().StringP("target", "t", "", "Optional CMake target name")

	rootCmd.SetFlagErrorFunc(FlagErrorFunc)
	rootCmd.AddCommand(list.ListCmd)
	return rootCmd
}

func FlagErrorFunc(cmd *cobra.Command, err error) error {
	if err != nil {
		log.Error(err)
		_ = cmd.Help()
	}
	return err
}
