/*
 * Copyright (c) 2023-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package csolution

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	builder "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder/cbuildidx"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder/cproject"
	utils "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type CSolutionBuilder struct {
	builder.BuilderParams
}

func (b CSolutionBuilder) formulateArgs(command []string) (args []string) {
	// formulate csolution arguments
	args = append(args, command...)

	if b.InputFile != "" {
		args = append(args, "--solution="+b.InputFile)
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
	if len(b.Options.Contexts) != 0 {
		for _, context := range b.Options.Contexts {
			args = append(args, "--context="+context)
		}
	}
	if b.Options.UseContextSet {
		args = append(args, "--context-set")
	}
	if b.Options.Toolchain != "" {
		args = append(args, "--toolchain="+b.Options.Toolchain)
	}
	if b.Options.Filter != "" {
		args = append(args, "--filter="+b.Options.Filter)
	}
	if b.Options.Verbose {
		args = append(args, "--verbose")
	}
	if b.Options.FrozenPacks {
		args = append(args, "--frozen-packs")
	}
	return
}

func (b CSolutionBuilder) runCSolution(args []string, quiet bool) (output string, err error) {
	csolutionBin, err := b.getCSolutionPath()
	if err != nil {
		return
	}

	// run csolution with args
	output, err = b.Runner.ExecuteCommand(csolutionBin, quiet, args...)
	return
}

func (b CSolutionBuilder) installMissingPacks() (err error) {
	if !b.Options.Packs {
		return nil
	}

	args := b.formulateArgs([]string{"list", "packs"})
	args = append(args, "-m")

	// Get list of missing packs
	output, err := b.runCSolution(args, false)
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

		// This call should be removed once the limitation of 'cpackget'
		// to handle '>=' in pack version, is resolved
		pack = utils.RemoveVersionRange(pack)

		args = []string{"add", pack, "--force-reinstall", "--agree-embedded-license", "--no-dependencies"}
		cpackgetBin := filepath.Join(b.InstallConfigs.BinPath, "cpackget"+b.InstallConfigs.BinExtn)
		if _, err := os.Stat(cpackgetBin); os.IsNotExist(err) {
			log.Error("error cpackget not found")
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

func (b CSolutionBuilder) generateBuildFiles() (err error) {
	args := b.formulateArgs([]string{"convert"})

	cbuildSetFile, err := b.getCbuildSetFilePath()
	if err != nil {
		return err
	}

	var selectedContexts []string
	if len(b.Options.Contexts) != 0 {
		selectedContexts = append(selectedContexts, b.Options.Contexts...)
	}

	_, err = os.Stat(cbuildSetFile)

	// Read contexts to be processed from cbuild-set file
	if b.Options.UseContextSet && err == nil {
		selectedContexts, _ = b.getSelectedContexts(cbuildSetFile)
	}

	// when using "cbuild setup *.csolution -S" with no existing cbuild-set file
	// Select first target-type and the first build-type for each project
	if b.Setup && b.Options.UseContextSet && (len(b.Options.Contexts) == 0) && errors.Is(err, os.ErrNotExist) {
		csolution, err := utils.ParseCSolutionFile(b.InputFile)
		if err != nil {
			return err
		}

		var buildType string
		if len(csolution.Solution.BuildTypes) > 0 {
			buildType = csolution.Solution.BuildTypes[0].Type
		} else {
			buildType = "*"
		}

		// Determine default context from the parsed solution file
		context := utils.ContextItem{
			ProjectName: "*",
			BuildType:   buildType,
			TargetType:  csolution.Solution.TargetTypes[0].Type,
		}

		// Create the default context
		defaultContext := utils.CreateContext(context)

		// Retrieve all available contexts in yml-order
		allContexts, err := b.listContexts(true, true)
		if err != nil {
			log.Error("error getting list of contexts: \"" + err.Error() + "\"")
			return err
		}

		// Ensure at least one context exists
		if len(allContexts) == 0 {
			return errors.New("error no context(s) found")
		}

		// Resolve the selected contexts including the default one
		selectedContexts, err = utils.ResolveContexts(allContexts, []string{defaultContext})
		if err != nil {
			return err
		}

		// Append selected contexts to the arguments
		for _, ctx := range selectedContexts {
			args = append(args, "--context="+ctx)
		}
	}

	_, err = b.runCSolution(args, false)

	// Execute this code exclusively upon invocation of the 'setup' command.
	// Its purpose is to update layer information within the *.cbuild-idx.yml files.
	if b.Setup {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 2 {
				args = []string{"list", "layers", b.InputFile, "--load=all", "--update-idx"}
				for _, context := range selectedContexts {
					args = append(args, "--context="+context)
				}
				_, listCmdErr := b.runCSolution(args, false)
				if listCmdErr != nil {
					err = listCmdErr
				}
			}
		}
	}

	return err
}

func (b CSolutionBuilder) getCprjFilePath(idxFile string, context string) (string, error) {
	var cprjPath string
	data, err := utils.ParseCbuildIndexFile(idxFile)
	if err == nil {
		var path string
		for _, cbuild := range data.BuildIdx.Cbuilds {
			if strings.Contains(strings.ToLower(cbuild.Cbuild), strings.ToLower(context)) {
				path = cbuild.Cbuild
				break
			}
		}
		if path == "" {
			err = errors.New("cprj file path not found")
		} else {
			cprjPath = filepath.Join(filepath.Dir(idxFile), filepath.Dir(path), context+".cprj")
		}
	}
	return cprjPath, err
}

func (b CSolutionBuilder) getSelectedContexts(filePath string) ([]string, error) {
	var contexts []string
	var retErr error

	if b.Options.UseContextSet {
		data, err := utils.ParseCbuildSetFile(filePath)
		if err == nil {
			for _, context := range data.ContextSet.Contexts {
				contexts = append(contexts, context.Context)
			}
		}
		retErr = err
	} else {
		data, err := utils.ParseCbuildIndexFile(filePath)
		if err == nil {
			for _, cbuild := range data.BuildIdx.Cbuilds {
				contexts = append(contexts, cbuild.Project+cbuild.Configuration)
			}
		}
		retErr = err
	}
	return contexts, retErr
}

func (b CSolutionBuilder) getCSolutionPath() (path string, err error) {
	path = filepath.Join(b.InstallConfigs.BinPath, "csolution"+b.InstallConfigs.BinExtn)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		log.Error("error csolution was not found: \"" + err.Error() + "\"")
	}
	return
}

func (b CSolutionBuilder) getIdxFilePath() (string, error) {
	projName, err := utils.GetProjectName(b.InputFile)
	if err != nil {
		return "", err
	}

	outputDir := b.Options.Output
	if outputDir == "" {
		outputDir = filepath.Dir(b.InputFile)
	}
	idxFilePath := utils.NormalizePath(filepath.Join(outputDir, projName+".cbuild-idx.yml"))
	return idxFilePath, nil
}

func (b CSolutionBuilder) getCbuildSetFilePath() (string, error) {
	projName, err := utils.GetProjectName(b.InputFile)
	if err != nil {
		return "", err
	}
	setFilePath := utils.NormalizePath(filepath.Join(filepath.Dir(b.InputFile), projName+".cbuild-set.yml"))

	return setFilePath, nil
}

func (b CSolutionBuilder) getProjsBuilders(selectedContexts []string) (projBuilders []builder.IBuilderInterface, err error) {
	buildOptions := b.Options

	// Set XML schema check to false, when input is yml
	if b.Options.Schema {
		buildOptions.Schema = false
	}

	idxFile, err := b.getIdxFilePath()
	if err != nil {
		return projBuilders, err
	}

	var projBuilder builder.IBuilderInterface
	if b.Options.UseCbuild2CMake {
		buildOptions.Contexts = selectedContexts
		// get idx builder
		projBuilder = cbuildidx.CbuildIdxBuilder{
			BuilderParams: builder.BuilderParams{
				Runner:         b.Runner,
				Options:        buildOptions,
				InputFile:      idxFile,
				InstallConfigs: b.InstallConfigs,
				Setup:          b.Setup,
			},
		}
		projBuilders = append(projBuilders, projBuilder)
	} else {
		for _, context := range selectedContexts {
			infoMsg := "Retrieve build information for context: \"" + context + "\""
			log.Info(infoMsg)

			// if --output is used, ignore provided --outdir and --intdir
			if b.Options.Output != "" && (b.Options.OutDir != "" || b.Options.IntDir != "") {
				log.Warn("output files are generated under: \"" +
					b.Options.Output + "\". Options --outdir and --intdir shall be ignored.")
			}

			cprjFile, err := b.getCprjFilePath(idxFile, context)
			if err != nil {
				log.Error("error getting cprj file: " + err.Error())
				return projBuilders, err
			}

			// get cprj builder
			projBuilder = cproject.CprjBuilder{
				BuilderParams: builder.BuilderParams{
					Runner:         b.Runner,
					Options:        buildOptions,
					InputFile:      cprjFile,
					InstallConfigs: b.InstallConfigs,
					Setup:          b.Setup,
				},
			}
			projBuilders = append(projBuilders, projBuilder)
		}
	}
	return projBuilders, err
}

func (b CSolutionBuilder) setBuilderOptions(builder *builder.IBuilderInterface, clean bool) {
	if b.Options.UseCbuild2CMake {
		idxBuilder := (*builder).(cbuildidx.CbuildIdxBuilder)
		idxBuilder.Options.Rebuild = false
		idxBuilder.Options.Clean = clean
		(*builder) = idxBuilder
	} else {
		cprjBuilder := (*builder).(cproject.CprjBuilder)
		cprjBuilder.Options.Rebuild = false
		cprjBuilder.Options.Clean = clean
		(*builder) = cprjBuilder
	}
}

func (b CSolutionBuilder) getBuilderInputFile(builder builder.IBuilderInterface) string {
	var inputFile string
	if b.Options.UseCbuild2CMake {
		idxBuilder := builder.(cbuildidx.CbuildIdxBuilder)
		inputFile = idxBuilder.InputFile
	} else {
		cprjBuilder := builder.(cproject.CprjBuilder)
		inputFile = cprjBuilder.InputFile
	}
	return inputFile
}

func (b CSolutionBuilder) cleanContexts(projBuilders []builder.IBuilderInterface) (err error) {
	for index := range projBuilders {
		b.setBuilderOptions(&projBuilders[index], true)
		err = projBuilders[index].Build()
		if err != nil {
			log.Error("error cleaning '" + b.getBuilderInputFile(projBuilders[index]) + "'")
		}
	}
	return
}

func (b CSolutionBuilder) buildContexts(selectedContexts []string, projBuilders []builder.IBuilderInterface) (err error) {
	operation := "Building"
	if b.Setup {
		operation = "Setting up"
	}

	for index := range projBuilders {
		var infoMsg string
		if b.Options.UseContextSet {
			infoMsg = operation + " \"" + selectedContexts[index] + "\""
		} else {
			progress := fmt.Sprintf("(%s/%d)", strconv.Itoa(index+1), len(selectedContexts))
			infoMsg = progress + " " + operation + " context: \"" + selectedContexts[index] + "\""
		}
		sep := strings.Repeat("=", len(infoMsg)+13) + "\n"
		_, _ = log.StandardLogger().Out.Write([]byte(sep))
		log.Info(infoMsg)

		b.setBuilderOptions(&projBuilders[index], false)

		err = projBuilders[index].Build()
		if err != nil {
			log.Error("error " + strings.ToLower(operation) + " '" + b.getBuilderInputFile(projBuilders[index]) + "'")
		}
	}
	return
}

func (b CSolutionBuilder) listContexts(quiet bool, ymlOrder bool) (contexts []string, err error) {
	args := b.formulateArgs([]string{"list", "contexts"})
	if ymlOrder {
		args = append(args, "--yml-order")
	}

	output, err := b.runCSolution(args, quiet)
	if err != nil {
		return
	}

	output = strings.ReplaceAll(output, " ", "")
	if output != "" {
		contexts = strings.Split(strings.ReplaceAll(strings.TrimSpace(output), "\r\n", "\n"), "\n")
	}
	return contexts, nil
}

func (b CSolutionBuilder) listToolchains(quiet bool) (toolchains []string, err error) {
	args := b.formulateArgs([]string{"list", "toolchains"})

	output, err := b.runCSolution(args, quiet)
	if err != nil {
		return
	}

	output = strings.ReplaceAll(output, " ", "")
	if output != "" {
		toolchains = strings.Split(strings.ReplaceAll(strings.TrimSpace(output), "\r\n", "\n"), "\n")
	}
	return toolchains, nil
}

func (b CSolutionBuilder) listEnvironment(quiet bool) (envConfigs []string, err error) {
	// get installed exe path and version number
	getInstalledExeInfo := func(name string) string {
		path, err := utils.GetInstalledExePath(name)
		if err != nil || path == "" {
			return "<Not Found>"
		}

		// run "exe --version" command
		versionStr, err := b.Runner.ExecuteCommand(path, true, "--version")
		if err != nil {
			versionStr = ""
		}

		// get version
		var version string
		if name == "cmake" {
			regex := "version\\s(.*?)\\s"
			re, err := regexp.Compile(regex)
			if err == nil {
				match := re.FindAllStringSubmatch(versionStr, 1)
				for index := range match {
					version = match[index][1]
					break
				}
			}
		} else {
			version = versionStr
		}
		info := path
		if version != "" {
			info += ", version " + version
		}
		return info
	}

	// step1: call csolution list environment
	args := []string{"list", "environment"}
	output, err := b.runCSolution(args, quiet)
	if err != nil {
		return
	}
	if output != "" {
		envConfigs = strings.Split(strings.ReplaceAll(strings.TrimSpace(output), "\r\n", "\n"), "\n")
	}

	// step2: add other environment info
	envConfigs = append(envConfigs, "cmake="+getInstalledExeInfo("cmake"))
	envConfigs = append(envConfigs, "ninja="+getInstalledExeInfo("ninja"))

	return envConfigs, nil
}

func (b CSolutionBuilder) ListContexts() error {
	_, err := b.listContexts(false, false)
	return err
}

func (b CSolutionBuilder) ListToolchains() error {
	_, err := b.listToolchains(false)
	return err
}

func (b CSolutionBuilder) ListEnvironment() error {
	envConfigs, err := b.listEnvironment(true)
	if err != nil {
		return err
	}
	for _, config := range envConfigs {
		_, _ = log.StandardLogger().Out.Write([]byte(config + "\n"))
	}
	return nil
}

func (b CSolutionBuilder) build() (err error) {
	var allContexts, selectedContexts []string
	if len(b.Options.Contexts) != 0 && !b.Options.UseContextSet {
		allContexts, err = b.listContexts(true, true)
		if err != nil {
			log.Error("error getting list of contexts: \"" + err.Error() + "\"")
			return err
		}
		selectedContexts, err = utils.ResolveContexts(allContexts, b.Options.Contexts)
	} else {
		var filePath string
		if b.Options.UseContextSet {
			filePath, err = b.getCbuildSetFilePath()
		} else {
			filePath, err = b.getIdxFilePath()
		}
		if err != nil {
			return err
		}
		selectedContexts, err = b.getSelectedContexts(filePath)
	}

	if err != nil {
		return err
	}

	totalContexts := strconv.Itoa(len(selectedContexts))
	log.Info("Processing " + totalContexts + " context(s)")

	// get builder for each selected context
	projBuilders, err := b.getProjsBuilders(selectedContexts)
	if err != nil {
		return err
	}

	// clean all selected contexts when rebuild or clean are requested
	if b.Options.Rebuild || b.Options.Clean {
		err = b.cleanContexts(projBuilders)
		if b.Options.Clean || err != nil {
			return err
		}
	}

	err = b.buildContexts(selectedContexts, projBuilders)
	return err
}

func (b CSolutionBuilder) Build() (err error) {
	_ = utils.UpdateEnvVars(b.InstallConfigs.BinPath, b.InstallConfigs.EtcPath)

	// STEP 1: Install missing pack(s)
	if err = b.installMissingPacks(); err != nil {
		log.Error("error installing missing packs")
		// Continue with build files generation upon setup command
		if !b.Setup {
			return err
		}
	}

	// STEP 2: Generate build file(s)
	if err = b.generateBuildFiles(); err != nil {
		log.Error("error generating build files")
		return err
	}

	// STEP 3: Build project(s)
	if err = b.build(); err != nil {
		log.Error("error building project(s)")
	}
	return err
}
