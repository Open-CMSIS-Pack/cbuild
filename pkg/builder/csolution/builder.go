/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package csolution

import (
	builder "cbuild/pkg/builder"
	"cbuild/pkg/builder/cproject"
	utils "cbuild/pkg/utils"
	"errors"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

type CSolutionBuilder struct {
	builder.BuilderParams
}

func (b CSolutionBuilder) listContexts(quite bool) (contexts []string, err error) {
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
		return nil, err
	}

	output, err := b.Runner.ExecuteCommand(csolutionBin, quite, args...)
	if err != nil {
		log.Error("error executing 'cbuild list contexts'")
		return nil, err
	}
	output = strings.ReplaceAll(output, " ", "")
	if output != "" {
		contexts = strings.Split(strings.ReplaceAll(strings.TrimSpace(output), "\r\n", "\n"), "\n")
	}
	return contexts, nil
}

func (b CSolutionBuilder) listToolchains(quite bool) (toolchains []string, err error) {
	csolutionBin := filepath.Join(b.InstallConfigs.BinPath, "csolution"+b.InstallConfigs.BinExtn)
	if _, err := os.Stat(csolutionBin); os.IsNotExist(err) {
		log.Error("csolution was not found")
		return toolchains, err
	}

	args := []string{"list", "toolchains"}
	if b.InputFile != "" {
		args = append(args, "--solution="+b.InputFile)
	}

	output, err := b.Runner.ExecuteCommand(csolutionBin, quite, args...)
	if err != nil {
		log.Error("error executing 'cbuild list toolchains'")
		return toolchains, err
	}
	output = strings.ReplaceAll(output, " ", "")
	if output != "" {
		toolchains = strings.Split(strings.ReplaceAll(strings.TrimSpace(output), "\r\n", "\n"), "\n")
	}
	return toolchains, nil
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
		pack = strings.ReplaceAll(pack, " ", "")
		if pack == "" {
			continue
		}
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

func (b CSolutionBuilder) ListContexts() error {
	_, err := b.listContexts(false)
	return err
}

func (b CSolutionBuilder) ListToolchains() error {
	_, err := b.listToolchains(false)
	return err
}

func (b CSolutionBuilder) getCPRJFilePath(idxFile string) (string, error) {
	var cprjPath string
	data, err := utils.ParseCbuildIndexFile(idxFile)
	if err == nil {
		var path string
		for _, cbuild := range data.BuildIdx.Cbuilds {
			if strings.Contains(cbuild.Cbuild, b.Options.Context) {
				path = cbuild.Cbuild
				break
			}
		}
		if path == "" {
			err = errors.New("cprj file path not found")
		} else {
			cprjPath = filepath.Dir(idxFile) + "/" + filepath.Dir(path) + "/" + b.Options.Context + ".cprj"
		}
	}
	return cprjPath, err
}

func (b CSolutionBuilder) Build() (err error) {
	_ = utils.UpdateEnvVars(b.InstallConfigs.BinPath, b.InstallConfigs.EtcPath)

	csolutionBin := filepath.Join(b.InstallConfigs.BinPath, "csolution"+b.InstallConfigs.BinExtn)
	if _, err := os.Stat(csolutionBin); os.IsNotExist(err) {
		log.Error("csolution was not found")
		return err
	}

	// validate context is empty
	if b.Options.Context == "" {
		contexts, err := b.listContexts(true)
		if err != nil {
			return err
		}

		if len(contexts) == 1 {
			b.Options.Context = contexts[0]
		} else {
			errMsg := "No context specified. One of the following contexts must be specified:\n" + strings.Join(contexts, "\n")
			return errors.New(errMsg)
		}
	}

	// install missing packs when --pack option is specified
	if b.Options.Packs {
		if err = b.installMissingPacks(); err != nil {
			return err
		}
	}

	nameTokens := strings.Split(filepath.Base(b.InputFile), ".")
	if len(nameTokens) != 3 {
		return errors.New("invalid csolution file name")
	}

	// formulate csolution arguments
	args := []string{
		"convert", "--solution=" + b.InputFile,
		"--context=" + b.Options.Context,
	}
	if b.Options.Output != "" {
		args = append(args, "--output="+b.Options.Output)
	}
	if b.Options.Load != "" {
		args = append(args, "--load="+b.Options.Load)
	}
	if !b.Options.Schema {
		args = append(args, "--no-check-schema")
	}
	if !b.Options.UpdateRte {
		args = append(args, "--no-update-rte")
	}

	// generate cprj files
	_, err = b.Runner.ExecuteCommand(csolutionBin, false, args...)
	if err != nil {
		log.Error("error building '" + b.InputFile + "'")
		return err
	}

	context, err := utils.ParseContext(b.Options.Context)
	if err != nil {
		return err
	}

	outputDir := b.Options.Output
	if outputDir == "" {
		outputDir = filepath.Dir(b.InputFile)
	}

	// get generated CPRJ file path from index yml
	cprjFile, err := b.getCPRJFilePath(outputDir + "/" + nameTokens[0] + ".cbuild-idx.yml")
	if err != nil {
		return err
	}

	b.Options.OutDir = filepath.Dir(cprjFile) + "/out/" + context.ProjectName + "/" +
		context.BuildType + "/" + context.TargetType
	b.Options.IntDir = filepath.Dir(cprjFile) + "/tmp/" + context.ProjectName + "/" +
		context.BuildType + "/" + context.TargetType

	// build generated CPRJ project
	cprjBuilder := cproject.CprjBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:         b.Runner,
			Options:        b.Options,
			InputFile:      cprjFile,
			InstallConfigs: b.InstallConfigs,
		},
	}

	return cprjBuilder.Build()
}
