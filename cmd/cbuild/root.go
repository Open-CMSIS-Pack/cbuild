/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"io"

	builder "cbuild/pkg/builder"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version string

func printVersion(file io.Writer) {
	fmt.Fprintf(file, "cbuild version %v\n", version)
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
	cobra.OnInitialize(initCobra)

	rootCmd := &cobra.Command{
		Use:           "cbuild <project.cprj> [flags]",
		Short:         "cbuild: Build Invocation " + version + " (C) 2022 ARM",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			versionFlag, _ := cmd.Flags().GetBool("version")
			if versionFlag {
				printVersion(cmd.OutOrStdout())
				return nil
			}

			helpFlag, _ := cmd.Flags().GetBool("help")
			if helpFlag {
				_ = cmd.Help()
				return nil
			}

			var cprjFile string
			if len(args) > 0 {
				cprjFile = args[0]
			} else {
				_ = cmd.Help()
				return nil
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
			clean, _ := cmd.Flags().GetBool("clean")

			cmdOptions := builder.CmdOptions{
				IntDir:    intDir,
				OutDir:    outDir,
				LockFile:  lockFile,
				LogFile:   logFile,
				Generator: generator,
				Target:    target,
				Jobs:      jobs,
				Quiet:     quiet,
				Debug:     debug,
				Clean:     clean,
			}

			err := builder.Build(cprjFile, cmdOptions)
			return err
		},
	}

	rootCmd.SetUsageTemplate(usageTemplate)

	rootCmd.Flags().BoolP("version", "v", false, "Print version")
	rootCmd.Flags().BoolP("help", "h", false, "Print usage")
	rootCmd.Flags().BoolP("quiet", "q", false, "Suppress output messages except build invocations")
	rootCmd.Flags().BoolP("debug", "d", false, "Enable debug messages")
	rootCmd.Flags().BoolP("clean", "c", false, "Remove intermediate and output directories")
	rootCmd.Flags().StringP("intdir", "i", "", "Set intermediate directory")
	rootCmd.Flags().StringP("outdir", "o", "", "Set output directory")
	rootCmd.Flags().StringP("update", "u", "", "Generate cprj file for reproducing current build")
	rootCmd.Flags().StringP("log", "l", "", "Save output messages in a log file")
	rootCmd.Flags().StringP("generator", "g", "Ninja", "Select build system generator")
	rootCmd.Flags().IntP("jobs", "j", 0, "Number of job slots for parallel execution")
	rootCmd.Flags().StringP("target", "t", "", "Optional CMake target name")

	rootCmd.SetFlagErrorFunc(FlagErrorFunc)

	return rootCmd
}

func FlagErrorFunc(cmd *cobra.Command, err error) error {
	if err != nil {
		log.Error(err)
		_ = cmd.Help()
	}
	return err
}

func initCobra() {
	viper.AutomaticEnv()
}
