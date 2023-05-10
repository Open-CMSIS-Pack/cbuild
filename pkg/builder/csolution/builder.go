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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type CSolutionBuilder struct {
	builder.BuilderParams
}

func (b CSolutionBuilder) formulateArgs(command []string) (args []string, err error) {
	if b.Options.Context != "" && b.Options.Configuration != "" {
		err = errors.New("options '--context' and '--configuration' cannot be used together")
		return
	}

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
	if b.Options.Configuration != "" {
		args = append(args, "--context=*"+b.Options.Configuration)
	}
	if b.Options.Context != "" {
		args = append(args, "--context="+b.Options.Context)
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
	args, err := b.formulateArgs([]string{"list", "packs"})
	if err != nil {
		log.Error("error in getting list of missing packs")
		return
	}
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
			cprjPath = filepath.Dir(idxFile) + "/" + filepath.Dir(path) + "/" + context + ".cprj"
		}
	}
	return cprjPath, err
}

func (b CSolutionBuilder) getCSolutionPath() (path string, err error) {
	path = filepath.Join(b.InstallConfigs.BinPath, "csolution"+b.InstallConfigs.BinExtn)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		log.Error("error csolution was not found: \"" + err.Error() + "\"")
	}
	return
}

func (b CSolutionBuilder) validateContext(allContexts []string, inputContext string) (context string, err error) {
	contextItem, err := utils.ParseContext(inputContext)
	if err != nil {
		return
	}

	context, err = utils.CreateContext(contextItem)
	if err != nil {
		return
	}
	if !utils.Contains(allContexts, context) {
		sort.Strings(allContexts)
		err = errors.New("specified context '" + inputContext +
			"' not found. One of the following contexts must be specified:\n" +
			strings.Join(allContexts, "\n"))
	}
	return
}

func (b CSolutionBuilder) processContext(context string, progress string) (err error) {
	infoMsg := progress + " Processing context: \"" + context + "\""
	fmt.Println(strings.Repeat("=", len(infoMsg)+13))
	log.Info(infoMsg)

	// if --output is used, ignore provided --outdir and --intdir
	if b.Options.Output != "" && (b.Options.OutDir != "" || b.Options.IntDir != "") {
		log.Warn("output files are generated under: \"" +
			b.Options.Output + "\". Options --outdir and --intdir shall be ignored.")
	}

	// get project name from file name
	nameTokens := strings.Split(filepath.Base(b.InputFile), ".")
	if len(nameTokens) != 3 {
		return errors.New("invalid csolution file name")
	}

	// get generated CPRJ file path from index yml
	outputDir := b.Options.Output
	if outputDir == "" {
		outputDir = filepath.Dir(b.InputFile)
	}
	cprjFile, err := b.getCprjFilePath(
		filepath.Join(outputDir, nameTokens[0]+".cbuild-idx.yml"), context)
	if err != nil {
		log.Error("error getting cprj file: " + err.Error())
		return err
	}

	// process generated CPRJ project
	cprjBuilder := cproject.CprjBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:         b.Runner,
			Options:        b.Options,
			InputFile:      cprjFile,
			InstallConfigs: b.InstallConfigs,
		},
	}
	err = cprjBuilder.Build()
	if err != nil {
		log.Error("error processing '" + cprjFile + "'")
	}
	return
}

func (b CSolutionBuilder) listConfigurations() (configurations []string, err error) {
	filter := b.Options.Filter
	b.Options.Filter = ""
	contexts, err := b.listContexts(true, false)
	if err != nil {
		err = errors.New("processing configurations list failed")
		return
	}

	// formulate solution contexts
	if len(contexts) != 0 {
		for _, context := range contexts {
			buildIdx := strings.Index(context, ".")
			targetIdx := strings.Index(context, "+")
			if buildIdx == -1 && targetIdx == -1 {
				continue
			}
			var config string
			if buildIdx != -1 {
				config = context[buildIdx:]
			} else {
				config = context[targetIdx:]
			}
			if filter != "" {
				if strings.Contains(config, filter) {
					configurations = utils.AppendUnique(configurations, config)
				}
				continue
			}
			configurations = utils.AppendUnique(configurations, config)
		}
	}

	if len(configurations) == 0 {
		if filter != "" {
			err = errors.New("no configuration was found with filter '" + filter + "'")
			return
		}
		err = errors.New("no configuration found")
		return
	}
	return configurations, nil
}

func (b CSolutionBuilder) listContexts(quiet bool, ymlOrder bool) (contexts []string, err error) {
	args, err := b.formulateArgs([]string{"list", "contexts"})
	if err != nil {
		return
	}

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
	args, err := b.formulateArgs([]string{"list", "toolchains"})
	if err != nil {
		return
	}

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

func (b CSolutionBuilder) ListConfigurations() error {
	configurations, err := b.listConfigurations()
	if err != nil {
		return err
	}
	fmt.Println(strings.Join(configurations, "\n"))
	return nil
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
		fmt.Println(config)
	}
	return nil
}

func (b CSolutionBuilder) Build() (err error) {
	_ = utils.UpdateEnvVars(b.InstallConfigs.BinPath, b.InstallConfigs.EtcPath)

	// get yml ordered list of all contexts
	allContexts, err := b.listContexts(true, true)
	if err != nil {
		log.Error("error getting list of contexts: \"" + err.Error() + "\"")
		return
	}

	// get list of contexts needs to be processed
	var selectedContexts []string
	if b.Options.Context != "" {
		var context string
		context, err = b.validateContext(allContexts, b.Options.Context)
		if err != nil {
			return
		}
		b.Options.Context = context
		selectedContexts = append(selectedContexts, context)
	} else {
		// get list of valid contexts from specified configuration
		selectedContexts, err = utils.GetSelectedContexts(allContexts, b.Options.Configuration)
		if err != nil {
			return
		}
	}

	args, err := b.formulateArgs([]string{"convert"})
	if err != nil {
		return
	}

	// install missing packs when --pack option is specified
	if b.Options.Packs {
		if err = b.installMissingPacks(); err != nil {
			log.Error("error installing missing packs: \"" + err.Error() + "\"")
			return err
		}
	}

	totalContexts := strconv.Itoa(len(selectedContexts))
	log.Info("Processing " + totalContexts + " context(s)")

	// step1: generate cprj files
	_, err = b.runCSolution(args, false)
	if err != nil {
		return err
	}

	// step2: process each selected context
	for index, context := range selectedContexts {
		progress := fmt.Sprintf("(%s/%s)", strconv.Itoa(index+1), totalContexts)
		err = b.processContext(context, progress)
		if err != nil {
			break
		}
	}
	return err
}
