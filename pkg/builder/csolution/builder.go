/*
 * Copyright (c) 2023-2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package csolution

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	builder "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder/cbuildidx"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder/cproject"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
	log "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/logger"
	utils "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
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
	if !b.Options.SchemaChk {
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
	if b.Options.Quiet {
		args = append(args, "--quiet")
	}
	if b.Options.UseTargetSet {
		args = append(args, "--active="+b.Options.TargetSet)
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
	args = append(args, "-m", "-q")

	if b.Setup {
		// If the setup command is triggered, skip the schema check when retrieving the list of missing packs.
		// This is because the schema check will be performed later during the YAML generation step.
		// Skipping it here avoids redundant checks and prevents potential double schema check errors.
		args = append(args, "-n")
	}

	// Get list of missing packs
	output, err := b.runCSolution(args, true)
	if err != nil {
		return err
	}

	// Installing missing packs
	missingPacks := strings.Split(strings.ReplaceAll(strings.TrimSpace(output), "\r\n", "\n"), "\n")
	for _, pack := range missingPacks {
		pack = strings.ReplaceAll(pack, " ", "")
		if pack == "" {
			continue
		}

		args = []string{"add", pack, "--force-reinstall", "--agree-embedded-license", "--no-dependencies"}
		cpackgetBin := filepath.Join(b.InstallConfigs.BinPath, "cpackget"+b.InstallConfigs.BinExtn)
		if _, err := os.Stat(cpackgetBin); os.IsNotExist(err) {
			return err
		}

		_, err = b.Runner.ExecuteCommand(cpackgetBin, false, args...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b CSolutionBuilder) generateBuildFiles() (err error) {
	args := b.formulateArgs([]string{"convert"})

	// when using "cbuild setup *.csolution.yml -S" with no existing cbuild-set file
	// Select first target-type and the first build-type for first listed project
	if b.Setup && b.Options.UseContextSet && (len(b.Options.Contexts) == 0) && errors.Is(err, os.ErrNotExist) {
		// Retrieve all available contexts in yml-order
		allYmlOrderedContexts, err := b.listContexts(true, true)
		if err != nil {
			return err
		}

		// Ensure at least one context exists
		if len(allYmlOrderedContexts) == 0 {
			return errutils.New(errutils.ErrNoContextFound)
		}

		// Append first context from yml ordered context list
		args = append(args, "--context="+allYmlOrderedContexts[0])
	}

	// on setup command, run csolution convert command with --quiet
	if b.Setup {
		args = append(args, "--quiet")
	}

	if !b.Options.UseCbuild2CMake {
		args = append(args, "--cbuildgen")
	}

	var stdErr string
	if b.Setup {
		var csolutionBin string
		csolutionBin, err = b.getCSolutionPath()
		if err != nil {
			return
		}
		log.Debug("csolution command: csolution " + strings.Join(args, " "))
		_, stdErr, err = utils.ExecuteCommand(csolutionBin, args...)
	} else {
		log.Debug("csolution command: csolution " + strings.Join(args, " "))

		//nolint:staticcheck // intentional logic for clarity
		_, convertErr := b.runCSolution(args, !b.Options.Debug && !b.Options.Verbose)
		if convertErr != nil {
			return convertErr
		}

		// Get index file
		idxFile, err := b.getIdxFilePath()
		if err != nil {
			return err
		}

		data, err := utils.ParseCbuildIndexFile(idxFile)
		if err != nil {
			return err
		}

		// Collect unique warnings
		uniqueWarnings := make(map[string]struct{})
		for _, cbuild := range data.BuildIdx.Cbuilds {
			for _, warning := range cbuild.Messages.Warnings {
				uniqueWarnings[warning] = struct{}{}
			}
		}

		// extract unregistered toolchains
		if len(uniqueWarnings) > 0 {
			var warnings []string
			for w := range uniqueWarnings {
				warnings = append(warnings, w)
			}

			outStr := strings.Join(warnings, ", ")
			re := regexp.MustCompile(`no compiler registered for '([^']+)'`)
			matches := re.FindAllStringSubmatch(outStr, -1)

			toolchains := make(map[string]struct{})
			for _, match := range matches {
				if len(match) > 1 {
					toolchains[match[1]] = struct{}{}
				}
			}

			if len(toolchains) > 0 {
				var missing []string
				for name := range toolchains {
					missing = append(missing, name)
				}
				return errutils.New(errutils.ErrNoCompilerRegistered, strings.Join(missing, ", "))
			}
		}

	}

	// Execute this code exclusively upon invocation of the 'setup' command.
	// Its purpose is to update layer information within the *.cbuild-idx.yml files.
	if b.Setup {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Added debug log for more info
			errCodeStr := fmt.Sprint(exitError.ExitCode())
			log.Debug("error code received: " + errCodeStr)

			if exitError.ExitCode() == 2 {
				args = b.formulateArgs([]string{"list", "layers", "--update-idx"})
				if !b.Options.Quiet {
					// force --quiet
					args = append(args, "--quiet")
				}
				_, listCmdErr := b.runCSolution(args, false)
				log.Debug("csolution command: csolution " + strings.Join(args, " "))
				if listCmdErr != nil {
					err = listCmdErr
				} else {
					reader := strings.NewReader(stdErr)
					scanner := bufio.NewScanner(reader)

					var errMsg string
					for scanner.Scan() {
						line := scanner.Text()
						if strings.Contains(line, "undefined variables in") || strings.HasSuffix(line, "-Layer$") {
							errMsg += (line + "\n")
						}
					}
					if errMsg != "" {
						errMsg += "To resolve undefined variables, copy the settings from cbuild-idx.yml to csolution.yml"
						utils.LogStdMsg(errMsg)
					}
				}
			} else {
				utils.LogStdMsg(stdErr)
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
			err = errutils.New(errutils.ErrCPRJNotFound, context+".cprj")
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
	_, err = os.Stat(path)
	return
}

func (b CSolutionBuilder) getIdxFilePath() (string, error) {
	projName := b.getProjectName(b.InputFile)
	outputDir := b.Options.Output
	if outputDir == "" {
		outputDir = filepath.Dir(b.InputFile)
	}
	idxFilePath := utils.NormalizePath(filepath.Join(outputDir, projName+".cbuild-idx.yml"))
	_, err := utils.FileExists(idxFilePath)
	if err != nil {
		return "", err
	}
	return idxFilePath, nil
}

func (b CSolutionBuilder) getProjectName(csolutionFile string) (projectName string) {
	csolutionFile = utils.NormalizePath(csolutionFile)
	nameTokens := strings.Split(filepath.Base(csolutionFile), ".")
	return nameTokens[0]
}

// This function merely constructs the path for the cbuild-set.yml file.
// It is the caller's responsibility to verify if the file exists."
func (b CSolutionBuilder) getCbuildSetFilePath() string {
	projName := b.getProjectName(b.InputFile)
	var cbuildSetDir string
	if len(b.Options.Output) > 0 {
		cbuildSetDir = b.Options.Output
	} else {
		cbuildSetDir = filepath.Dir(b.InputFile)
	}
	return utils.NormalizePath(filepath.Join(cbuildSetDir, projName+".cbuild-set.yml"))
}

func (b CSolutionBuilder) getProjsBuilders(selectedContexts []string) (projBuilders []builder.IBuilderInterface, err error) {
	buildOptions := b.Options

	// Disable XML schema check if the input is YAML
	if b.Options.SchemaChk {
		buildOptions.SchemaChk = false
	}

	// Get the index file path
	idxFile, err := b.getIdxFilePath()
	if err != nil {
		return nil, err
	}

	var projBuilder builder.IBuilderInterface
	for _, context := range selectedContexts {
		infoMsg := "Retrieve build information for context: \"" + context + "\""
		log.Info(infoMsg)

		// Warn if --output is used along with --outdir or --intdir
		if b.Options.Output != "" && (b.Options.OutDir != "" || b.Options.IntDir != "") {
			log.Warn("output files are generated under: \"" +
				b.Options.Output + "\". Options --outdir and --intdir shall be ignored.")
		}

		if b.Options.UseCbuild2CMake {
			// Create a builder for cbuild2CMake
			projBuilder = cbuildidx.CbuildIdxBuilder{
				BuilderParams: builder.BuilderParams{
					Runner:         b.Runner,
					Options:        buildOptions,
					InputFile:      idxFile,
					InstallConfigs: b.InstallConfigs,
					Setup:          b.Setup,
					BuildContext:   context,
				},
			}
		} else {
			// Get the .cprj file path for the current context
			cprjFile, err := b.getCprjFilePath(idxFile, context)
			if err != nil {
				return projBuilders, err
			}

			// Create a builder for cproject
			projBuilder = cproject.CprjBuilder{
				BuilderParams: builder.BuilderParams{
					Runner:         b.Runner,
					Options:        buildOptions,
					InputFile:      cprjFile,
					InstallConfigs: b.InstallConfigs,
					Setup:          b.Setup,
				},
			}
		}

		// Append the created builder to the list
		projBuilders = append(projBuilders, projBuilder)
	}
	return projBuilders, nil
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

func (b CSolutionBuilder) buildContexts(selectedContexts []string, projBuilders []builder.IBuilderInterface) (err error) {
	operation := "Building"
	if b.Setup {
		operation = "Setting up"
	}

	buildPassCnt := 0
	buildFailCnt := 0
	var totalBuildTime time.Duration
	for index := range projBuilders {
		progress := fmt.Sprintf("(%s/%d)", strconv.Itoa(index+1), len(selectedContexts))
		buildMsg := progress + " " + operation + " context: \"" + selectedContexts[index] + "\""

		utils.PrintSeparator("-", len(buildMsg))
		utils.LogStdMsg(buildMsg)
		b.setBuilderOptions(&projBuilders[index], false)

		buildStartTime := time.Now()
		buildErr := projBuilders[index].Build()
		if buildErr != nil {
			err = buildErr
			buildFailCnt += 1
		} else {
			buildPassCnt += 1
		}
		buildEndTime := time.Now()
		elapsedTime := buildEndTime.Sub(buildStartTime)
		totalBuildTime += elapsedTime
	}
	if !b.Setup {
		buildSummary := fmt.Sprintf("Build summary: %d succeeded, %d failed - Time Elapsed: %s", buildPassCnt, buildFailCnt, utils.FormatTime(totalBuildTime))
		sepLen := len(buildSummary)
		utils.PrintSeparator("-", sepLen)
		utils.LogStdMsg(buildSummary)
		utils.PrintSeparator("=", sepLen)
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
		utils.LogStdMsg(config)
	}
	return nil
}

func (b CSolutionBuilder) build() (err error) {
	var allContexts, selectedContexts []string
	if len(b.Options.Contexts) != 0 && !b.Options.UseContextSet {
		allContexts, err = b.listContexts(true, true)
		if err != nil {
			log.Error(err)
			return err
		}
		selectedContexts, err = utils.ResolveContexts(allContexts, b.Options.Contexts)
	} else {
		var filePath string
		if b.Options.UseContextSet {
			filePath = b.getCbuildSetFilePath()
			_, err = utils.FileExists(filePath)
		} else {
			filePath, err = b.getIdxFilePath()
		}
		if err != nil {
			log.Error(err)
			return err
		}
		selectedContexts, err = b.getSelectedContexts(filePath)
	}

	if err != nil {
		log.Error(err)
		return err
	}

	totalContexts := strconv.Itoa(len(selectedContexts))
	log.Info("Processing " + totalContexts + " context(s)")

	// get builder for each selected context
	projBuilders, err := b.getProjsBuilders(selectedContexts)
	if err != nil {
		log.Error(err)
		return err
	}

	needRebuild, err := b.needRebuild()
	if err != nil {
		log.Error(err)
		return err
	}
	if needRebuild {
		// Perform the clean operation
		err := b.Clean()
		if err != nil {
			log.Error(err)
			return err
		}
	}

	if len(b.Options.Target) == 0 {
		err = b.buildContexts(selectedContexts, projBuilders)
	} else {
		// build only cmake target when --target is specified
		err = projBuilders[0].Build()
	}
	return err
}

func (b CSolutionBuilder) Build() (err error) {
	_ = utils.UpdateEnvVars(b.InstallConfigs.BinPath, b.InstallConfigs.EtcPath)

	// STEP 1: Install missing pack(s)
	if err = b.installMissingPacks(); err != nil {
		// Continue with build files generation upon setup command
		if !b.Setup {
			log.Error(err)
			return err
		}
	}

	// STEP 2: Generate build file(s)
	if err = b.generateBuildFiles(); err != nil {
		log.Error(err)
		return err
	}

	// STEP 3: Build project(s)
	return b.build()
}

func (b CSolutionBuilder) needRebuild() (bool, error) {
	// Rebuild was already requested by user
	if b.Options.Rebuild || !b.Options.UseCbuild2CMake {
		return false, nil
	}

	// Check if the project is moved or renamed or tmp dir is changed
	if b.isProjectMoved() {
		return true, nil
	}

	// Get cbuild-idx file path
	idxFilePath, err := b.getIdxFilePath()
	if err != nil {
		return false, err
	}

	// Read .cbuild-idx.yml and check if "rebuild" node exist
	rebuild, err := b.hasRebuildNode(idxFilePath)
	if err != nil {
		return false, err
	}
	return rebuild, nil
}

func (b CSolutionBuilder) isProjectMoved() bool {
	csolutionAbsPath, _ := filepath.Abs(b.InputFile)
	rootPath := filepath.Dir(csolutionAbsPath)
	intDirPath := filepath.Join(rootPath, "tmp")
	cmakeCacheFile := filepath.Join(intDirPath, "CMakeCache.txt")

	// check if input file exists
	_, err := utils.FileExists(cmakeCacheFile)
	if err != nil {
		// File doesn't exist, rebuild not needed
		return false
	}

	file, err := os.Open(cmakeCacheFile)
	if err != nil {
		return true
	}
	defer file.Close()

	// Initialize a scanner to read file
	scanner := bufio.NewScanner(file)
	prefixStr := "CMAKE_CACHEFILE_DIR:INTERNAL="

	// Search for prefixStr
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, prefixStr) {
			path := strings.TrimPrefix(line, prefixStr)
			equal, _ := utils.ComparePaths(path, intDirPath)
			if equal {
				return false // paths match, rebuild not needed
			} else {
				return true // paths do not match, rebuild needed
			}
		}
	}

	// prefixStr was not found in the file, rebuild needed
	return true
}

// hasRebuildNode checks if there is any rebuild required based on the given index file path.
// It returns true if a rebuild is needed, otherwise false.
func (b CSolutionBuilder) hasRebuildNode(idxFilePath string) (bool, error) {
	// Read the cbuild-idx file
	data, err := utils.ParseCbuildIndexFile(idxFilePath)
	if err != nil {
		return false, err
	}

	// Check if the main build index requires a rebuild
	if data.BuildIdx.Rebuild {
		return true, nil
	}

	// Check if any of the contexts requires a rebuild
	for _, cbuild := range data.BuildIdx.Cbuilds {
		if cbuild.Rebuild {
			return true, nil
		}
	}

	return false, nil
}

func (b CSolutionBuilder) Clean() (err error) {
	// Get list of cleanable contexts
	cleanableContexts, err := b.getContextsToClean()
	if err != nil {
		return err
	}

	outputDir := b.Options.Output

	var tmpDir string
	if !b.Options.UseCbuild2CMake {
		// Use default path when --cbuildgen option is used
		tmpDir = filepath.Join(filepath.Dir(b.InputFile), outputDir, "tmp")
	} else {
		tmpDir, err = utils.GetTmpDir(b.InputFile, outputDir)
		if err != nil {
			return err
		}
	}

	// Clean tmp dir
	if err := utils.DeleteAll(tmpDir, []string{}); err != nil {
		if !b.Options.Clean {
			log.Warn(err.Error())
		}
	}

	// Clean out dir
	var seplen int
	for index, context := range cleanableContexts {
		progress := fmt.Sprintf("(%s/%d)", strconv.Itoa(index+1), len(cleanableContexts))
		cleanMsg := progress + " Cleaning context: \"" + cleanableContexts[index] + "\""
		if seplen == 0 {
			seplen = len(cleanMsg)
			utils.PrintSeparator("-", seplen)
		}
		utils.LogStdMsg(cleanMsg)

		idxFile, err := b.getIdxFilePath()
		if err == nil {
			outDir, err := utils.GetOutDir(idxFile, context)
			if err != nil {
				log.Error("error cleaning '" + context + "'")
			}
			if err = utils.DeleteAll(outDir, []string{"*.cbuild.yml", "*.cbuild-run.yml"}); err != nil {
				if !b.Options.Clean {
					log.Warn(err.Error())
				}
			}
		}
	}

	if b.Options.Clean {
		utils.PrintSeparator("-", seplen)
	}
	log.Info("clean finished successfully!")
	return nil
}

func (b *CSolutionBuilder) getContextsToClean() (contexts []string, err error) {
	// Retrieve all available contexts
	builder := b
	builder.Options.SchemaChk = false
	allContexts, err := builder.listContexts(true, true)
	if err != nil {
		return []string{}, err
	}

	hasContextOption := (len(b.Options.Contexts) > 0 && !b.Options.UseContextSet)
	hasTargetSetOption := (b.Options.TargetSet != "")

	if hasTargetSetOption && (len(b.Options.Contexts) > 0 || b.Options.UseContextSet) {
		err := errutils.New(errutils.ErrInvalidTargetSetUsage)
		return []string{}, err
	}

	if hasContextOption || hasTargetSetOption {
		var contextInputs []string

		if hasTargetSetOption {
			// Use project contexts associated with the specified target-set
			contextInputs = utils.GetTargetSetProjectContexts(b.InputFile, b.Options.TargetSet)
		} else if hasContextOption {
			// Use the explicitly provided contexts
			contextInputs = b.Options.Contexts
		}

		// Resolve contexts if inputs are available
		contexts, err = utils.ResolveContexts(allContexts, contextInputs)
		if err != nil {
			return []string{}, err
		}
		return contexts, nil
	}

	// Handle context selection from a cbuild set file
	if b.Options.UseContextSet {
		filePath := b.getCbuildSetFilePath()
		if exists, err := utils.FileExists(filePath); err != nil || !exists {
			return []string{}, err
		}

		contexts, err = b.getSelectedContexts(filePath)
		if err != nil {
			return []string{}, err
		}
		return contexts, nil
	}

	// Default to using all available contexts
	return allContexts, nil
}

func (b CSolutionBuilder) ListTargetSets() error {
	_, err := b.listTargetSets(false)
	return err
}

func (b CSolutionBuilder) listTargetSets(quiet bool) (targetSets []string, err error) {
	args := b.formulateArgs([]string{"list", "target-sets"})

	output, err := b.runCSolution(args, quiet)
	if err != nil {
		return
	}

	if output != "" {
		targetSets = strings.Split(strings.ReplaceAll(strings.TrimSpace(output), "\r\n", "\n"), "\n")
	}
	return targetSets, nil
}
