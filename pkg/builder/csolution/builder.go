/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package csolution

import (
	builder "cbuild/pkg/builder"
	"cbuild/pkg/builder/cproject"
	utils "cbuild/pkg/utils"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

type CSolutionBuilder struct {
	builder.BuilderParams
}

func (b CSolutionBuilder) ListContexts() (err error) {
	args := []string{"list", "contexts", "--solution=" + b.InputFile}

	if b.Options.Filter != "" {
		args = append(args, "--filter="+b.Options.Filter)
	}

	if !b.Options.Schema {
		args = append(args, "--no-check-schema")
	}

	csolutionBin := filepath.Join(b.InstallConfigs.BinPath, "csolution"+b.InstallConfigs.BinExtn)
	if _, err := os.Stat(csolutionBin); os.IsNotExist(err) {
		log.Error("csolution was not found")
		return err
	}

	_, err = b.Runner.ExecuteCommand(csolutionBin, false, args...)
	if err != nil {
		log.Error("error executing 'cbuild list-contexts'")
		return err
	}
	return nil
}

func (b CSolutionBuilder) ListToolchains() (err error) {
	csolutionBin := filepath.Join(b.InstallConfigs.BinPath, "csolution"+b.InstallConfigs.BinExtn)
	if _, err := os.Stat(csolutionBin); os.IsNotExist(err) {
		log.Error("csolution was not found")
		return err
	}

	args := []string{"list", "toolchains", "--solution=" + b.InputFile}

	_, err = b.Runner.ExecuteCommand(csolutionBin, false, args...)
	if err != nil {
		log.Error("error executing 'cbuild list-toolchains'")
		return err
	}
	return nil
}

func (b CSolutionBuilder) Build() (err error) {
	_ = utils.UpdateEnvVars(b.InstallConfigs.BinPath, b.InstallConfigs.ETCPath)

	if b.Options.Packs {
		if err = b.installMissingPacks(); err != nil {
			return err
		}
	}

	outDir := b.Options.OutDir
	if outDir == "" {
		outDir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	args := []string{
		"convert", "--solution=" + b.InputFile, "--context=" + b.Options.Context,
		"--output=" + outDir, "--load=" + b.Options.Load}
	if !b.Options.Schema {
		args = append(args, "-n")
	}

	csolutionBin := filepath.Join(b.InstallConfigs.BinPath, "csolution"+b.InstallConfigs.BinExtn)
	if _, err := os.Stat(csolutionBin); os.IsNotExist(err) {
		log.Error("csolution was not found")
		return err
	}

	_, err = b.Runner.ExecuteCommand(csolutionBin, false, args...)
	if err != nil {
		log.Error("error building '" + b.InputFile + "'")
		return err
	}

	// build generated CPRJ project
	cprjBuilder := cproject.CPRJBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:         b.Runner,
			Options:        b.Options,
			InputFile:      outDir + "/" + b.Options.Context + ".cprj",
			InstallConfigs: b.InstallConfigs,
		},
	}

	return cprjBuilder.Build()
}

func (b CSolutionBuilder) installMissingPacks() (err error) {
	args := []string{"list", "packs", "--solution=" + b.InputFile, "-m",
		"--context=" + b.Options.Context, "--filter=" + b.Options.Filter}

	if !b.Options.Schema {
		args = append(args, "--no-check-schema")
	}

	// Get list of missing packs
	csolutionBin := filepath.Join(b.InstallConfigs.BinPath, "csolution"+b.InstallConfigs.BinExtn)
	if _, err := os.Stat(csolutionBin); os.IsNotExist(err) {
		log.Error("csolution was not found")
		return err
	}

	output, err := b.Runner.ExecuteCommand(csolutionBin, false, args...)
	if err != nil {
		log.Error("error in getting list of missing packs")
		return err
	}

	// Installing missing packs
	missingPacks := strings.Split(strings.ReplaceAll(strings.TrimSpace(output), "\r\n", "\n"), "\n")
	for _, pack := range missingPacks {
		args = []string{"pack", "add", pack, "--force-reinstall", "--agree-embedded-license"}
		cpackgetBin := filepath.Join(b.InstallConfigs.BinPath, "cpackget"+b.InstallConfigs.BinExtn)
		if _, err := os.Stat(cpackgetBin); os.IsNotExist(err) {
			log.Error("cpackget was not found")
			return err
		}

		_, err = b.Runner.ExecuteCommand(cpackgetBin, false, args...)
		if err != nil {
			log.Error("error installing pack : " + pack)
			return err
		}
	}

	return nil
}
