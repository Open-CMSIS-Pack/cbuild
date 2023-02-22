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
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

type CSolutionBuilder struct {
	builder.BuilderParams
}

func (b CSolutionBuilder) listConfigurations() (configurations []string, err error) {
	filter := b.Options.Filter
	b.Options.Filter = ""
	contexts, err := b.listContexts(true, false)
	if err != nil {
		return configurations, errors.New("processing configurations list failed")
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
			log.Error("no configuration was found with filter '" + filter + "'")
			return configurations, errors.New("processing configurations list failed")
		}
		log.Info("no configuration found")
	}
	return configurations, nil
}

func (b CSolutionBuilder) listContexts(quite bool, ymlOrder bool) (contexts []string, err error) {
	args := []string{"list", "contexts", "--solution=" + b.InputFile}

	if b.Options.Filter != "" {
		args = append(args, "--filter="+b.Options.Filter)
	}

	if !b.Options.Schema {
		args = append(args, "--no-check-schema")
	}

	if ymlOrder {
		args = append(args, "--yml-order")
	}

	csolutionBin, err := b.getCsolutionPath()
	if err != nil {
		return
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
	csolutionBin, err := b.getCsolutionPath()
	if err != nil {
		return
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

	csolutionBin, err := b.getCsolutionPath()
	if err != nil {
		return
	}

	// Get list of missing packs
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

func (b CSolutionBuilder) getCprjFilePath(idxFile string, context string) (string, error) {
	var cprjPath string
	data, err := utils.ParseCbuildIndexFile(idxFile)
	if err == nil {
		var path string
		for _, cbuild := range data.BuildIdx.Cbuilds {
			if strings.Contains(cbuild.Cbuild, context) {
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

func (b CSolutionBuilder) getCsolutionPath() (path string, err error) {
	path = filepath.Join(b.InstallConfigs.BinPath, "csolution"+b.InstallConfigs.BinExtn)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		log.Error("error csolution was not found: \"" + err.Error() + "\"")
	}
	return
}

func (b CSolutionBuilder) validateContext(allContexts []string) (err error) {
	_, err = utils.ParseContext(b.Options.Context)
	if err != nil {
		return
	}

	if !utils.Contains(allContexts, b.Options.Context) {
		sort.Strings(allContexts)
		err = errors.New("specified context '" + b.Options.Context +
			"' not found. One of the following contexts must be specified:\n" +
			strings.Join(allContexts, "\n"))
	}
	return
}

func (b CSolutionBuilder) Build() (err error) {
	_ = utils.UpdateEnvVars(b.InstallConfigs.BinPath, b.InstallConfigs.EtcPath)
	csolutionBin, err := b.getCsolutionPath()
	if err != nil {
		return
	}

	if b.Options.Context != "" && b.Options.Configuration != "" {
		err = errors.New("options '--context' and '--configuration' cannot be used together")
		return
	}

	// get yml ordered list of all contexts
	allContexts, err := b.listContexts(true, true)
	if err != nil {
		log.Error("error getting list of contexts: \"" + err.Error() + "\"")
		return
	}

	var selectedContexts []string
	if b.Options.Context != "" {
		if err = b.validateContext(allContexts); err != nil {
			return
		}
		selectedContexts = append(selectedContexts, b.Options.Context)
	} else {
		if b.Options.Configuration == "" {
			// build all contexts when configuration is empty
			selectedContexts = allContexts
		} else {
			// get list of valid contexts from specified configuration
			selectedContexts, err = utils.GetSelectedContexts(allContexts, b.Options.Configuration)
			if err != nil {
				return
			}
		}
	}

	// install missing packs when --pack option is specified
	if b.Options.Packs {
		if err = b.installMissingPacks(); err != nil {
			log.Error("error installing missing packs: \"" + err.Error() + "\"")
			return err
		}
	}

	nameTokens := strings.Split(filepath.Base(b.InputFile), ".")
	if len(nameTokens) != 3 {
		return errors.New("invalid csolution file name")
	}

	var formulatePath = func(cprjFilePath string, dir string, context utils.ContextItem) string {
		path := filepath.Join(filepath.Dir(cprjFilePath), dir, context.ProjectName)
		if context.BuildType != "" {
			path = filepath.Join(path, context.BuildType)
		}
		path = filepath.Join(path, context.TargetType)
		return path
	}

	// formulate csolution arguments
	args := []string{"convert", "--solution=" + b.InputFile}
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
		configurationItem, err := utils.ParseConfiguration(b.Options.Configuration)
		if err != nil {
			return err
		}
		contextQuery := "*" + utils.CreateConfiguration(configurationItem)
		if configurationItem.TargetType == "" {
			contextQuery += "*"
		}
		args = append(args, "--context="+contextQuery)
	}
	if b.Options.Context != "" {
		args = append(args, "--context="+b.Options.Context)
	}

	// step1: generate cprj files
	_, err = b.Runner.ExecuteCommand(csolutionBin, false, args...)
	if err != nil {
		log.Error(err)
		return err
	}

	// build each selected context
	for _, context := range selectedContexts {
		infoMsg := "Building context: \"" + context + "\""
		separator := strings.Repeat("=", len(infoMsg)+13)
		log.Info(infoMsg + "\n" + separator)

		// if --output is used, ignore provided --outdir and --intdir
		if b.Options.Output != "" && (b.Options.OutDir != "" || b.Options.IntDir != "") {
			log.Warn("output files are generated under: \"" + b.Options.Output + "\". Options --outdir and --intdir shall be ignored.")
		}

		// step2: get generated CPRJ file path from index yml
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

		// step3: formulate outdir & intdir path
		selectedContext, _ := utils.ParseContext(context)
		cprjBuildOptions := b.Options
		cprjBuildOptions.OutDir = formulatePath(cprjFile, "out", selectedContext)
		cprjBuildOptions.IntDir = formulatePath(cprjFile, "tmp", selectedContext)

		log.Debug("outdir: " + b.Options.OutDir)
		log.Debug("intdir: " + b.Options.IntDir)

		// step4: build generated CPRJ project
		cprjBuilder := cproject.CprjBuilder{
			BuilderParams: builder.BuilderParams{
				Runner:         b.Runner,
				Options:        cprjBuildOptions,
				InputFile:      cprjFile,
				InstallConfigs: b.InstallConfigs,
			},
		}
		if err = cprjBuilder.Build(); err != nil {
			log.Error("error building '" + cprjFile + "'")
			break
		}
	}
	return err
}
