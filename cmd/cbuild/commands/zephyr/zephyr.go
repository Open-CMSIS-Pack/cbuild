/*
 * Copyright (c) 2026 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package zephyr

import (
	"os"
	"path/filepath"
	"strings"

	builder "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	csolution "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder/csolution"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
	log "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/logger"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	"github.com/spf13/cobra"
)

const (
	defaultToolchain = "GCC"
	defaultDevice    = "ARMCM55"
	defaultPack      = "ARM::Cortex_DFP"
)

var (
	getInstallConfigs = utils.GetInstallConfigs
	runnerInterface   = func() utils.RunnerInterface { return utils.Runner{} }
)

func createSolutionContent(moduleName string, toolchain string, device string, packs []string) string {
	var builder strings.Builder
	builder.WriteString("solution:\n\n")
	builder.WriteString("  compiler: ")
	builder.WriteString(toolchain)
	builder.WriteString("\n\n")
	builder.WriteString("  packs:\n")
	for _, pack := range packs {
		builder.WriteString("    - pack: ")
		builder.WriteString(pack)
		builder.WriteString("\n")
	}
	builder.WriteString("\n")
	builder.WriteString("  target-types:\n")
	builder.WriteString("    - type: ")
	builder.WriteString(device)
	builder.WriteString("\n")
	builder.WriteString("      device: ")
	builder.WriteString(device)
	builder.WriteString("\n\n")
	builder.WriteString("  projects:\n")
	builder.WriteString("    - project: ")
	builder.WriteString(moduleName)
	builder.WriteString(".cproject.yml\n")

	return builder.String()
}

func createProjectContent(clayers []string) string {
	var builder strings.Builder
	builder.WriteString("project:\n\n")
	builder.WriteString("  layers:\n")

	for _, clayer := range clayers {
		builder.WriteString("    - layer: ")
		builder.WriteString(clayer)
		builder.WriteString("\n")
	}

	return builder.String()
}

func normalizeLayers(input []string) []string {
	layers := make([]string, 0, len(input))
	for _, item := range input {
		layer := strings.TrimSpace(item)
		if layer == "" {
			continue
		}
		layers = append(layers, layer)
	}
	return layers
}

func isSimpleModuleName(moduleName string) bool {
	if moduleName == "." || moduleName == ".." {
		return false
	}
	if strings.ContainsAny(moduleName, `/\`) {
		return false
	}
	return moduleName == filepath.Base(moduleName)
}

func resolveAndValidateLayers(input []string, cwd string) ([]string, error) {
	layers := normalizeLayers(input)
	resolvedLayers := make([]string, 0, len(layers))

	for _, layer := range layers {
		layerPath := layer
		if !filepath.IsAbs(layerPath) {
			layerPath = filepath.Join(cwd, layerPath)
		}

		absolutePath, err := filepath.Abs(layerPath)
		if err != nil {
			return nil, errutils.New(errutils.ErrInvalidClayerPath, layer, err)
		}

		if !strings.HasSuffix(strings.ToLower(absolutePath), ".clayer.yml") {
			return nil, errutils.New(errutils.ErrInvalidFileExtension, layer, "*.clayer.yml")
		}

		info, err := os.Stat(absolutePath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, errutils.New(errutils.ErrFileNotExist, absolutePath)
			}
			return nil, errutils.New(errutils.ErrClayerAccess, absolutePath, err)
		}

		if info.IsDir() {
			return nil, errutils.New(errutils.ErrClayerPathIsDirectory, absolutePath)
		}

		resolvedLayers = append(resolvedLayers, absolutePath)
	}

	return resolvedLayers, nil
}

func makeLayersRelativeToProject(input []string, projectFile string) ([]string, error) {
	projectDir := filepath.Dir(projectFile)
	relativeLayers := make([]string, 0, len(input))

	for _, layer := range input {
		relativeLayer, err := filepath.Rel(projectDir, layer)
		if err != nil {
			return nil, errutils.New(errutils.ErrRelativizeClayerPath, layer, projectFile, err)
		}
		relativeLayers = append(relativeLayers, filepath.ToSlash(relativeLayer))
	}

	return relativeLayers, nil
}

func getToolBinary(binPath string, name string, ext string) (string, error) {
	toolPath := filepath.Join(binPath, name+ext)
	if _, err := os.Stat(toolPath); err != nil {
		return "", err
	}
	return toolPath, nil
}

func generateZephyrModule(cmd *cobra.Command, _ []string) error {
	moduleName, _ := cmd.Flags().GetString("module")
	toolchain, _ := cmd.Flags().GetString("toolchain")
	device, _ := cmd.Flags().GetString("device")
	clayers, _ := cmd.Flags().GetStringSlice("clayer")
	packs, _ := cmd.Flags().GetStringSlice("packs")
	originalDir, err := os.Getwd()
	if err != nil {
		log.Error(err)
		return err
	}

	moduleName = strings.TrimSpace(moduleName)
	if moduleName == "" {
		err := errutils.New(errutils.ErrMissingModuleArg)
		log.Error(err)
		return err
	}
	if !isSimpleModuleName(moduleName) {
		err := errutils.New(errutils.ErrInvalidInputArg, "--module")
		log.Error(err)
		return err
	}

	validLayers, err := resolveAndValidateLayers(clayers, originalDir)
	if err != nil {
		log.Error(err)
		return err
	}
	if len(validLayers) == 0 {
		err := errutils.New(errutils.ErrMissingClayerArg)
		log.Error(err)
		return err
	}

	validPacks := normalizeLayers(packs)
	if len(validPacks) == 0 {
		validPacks = []string{defaultPack}
	}

	tmpDir := filepath.Join(originalDir, "..", moduleName+"-tmp")
	if err = os.RemoveAll(tmpDir); err != nil {
		log.Error(err)
		return err
	}
	if err = os.MkdirAll(tmpDir, 0700); err != nil {
		log.Error(err)
		return err
	}
	defer os.RemoveAll(tmpDir)

	if err = os.Chdir(tmpDir); err != nil {
		log.Error(err)
		return err
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	solutionFile := moduleName + ".csolution.yml"
	projectFile := moduleName + ".cproject.yml"
	projectPath := filepath.Join(tmpDir, projectFile)

	projectLayers, err := makeLayersRelativeToProject(validLayers, projectPath)
	if err != nil {
		log.Error(err)
		return err
	}

	solutionContent := createSolutionContent(moduleName, toolchain, device, validPacks)
	projectContent := createProjectContent(projectLayers)

	if err = os.WriteFile(solutionFile, []byte(solutionContent), 0600); err != nil {
		log.Error(err)
		return err
	}
	if err = os.WriteFile(projectFile, []byte(projectContent), 0600); err != nil {
		log.Error(err)
		return err
	}

	configs, err := getInstallConfigs()
	if err != nil {
		log.Error(err)
		return err
	}

	csolutionBin, err := getToolBinary(configs.BinPath, "csolution", configs.BinExtn)
	if err != nil {
		log.Error(err)
		return err
	}

	cbuild2cmakeBin, err := getToolBinary(configs.BinPath, "cbuild2cmake", configs.BinExtn)
	if err != nil {
		log.Error(err)
		return err
	}

	runner := runnerInterface()
	csolutionBuilder := csolution.CSolutionBuilder{
		BuilderParams: builder.BuilderParams{
			Runner:         runner,
			Options:        builder.Options{Packs: true, SchemaChk: true, UpdateRte: true},
			InputFile:      solutionFile,
			InstallConfigs: configs,
		},
	}
	if err = csolutionBuilder.InstallMissingPacks(); err != nil {
		log.Error(err)
		return err
	}

	_, err = runner.ExecuteCommand(csolutionBin, false, "convert", solutionFile)
	if err != nil {
		log.Error(err)
		return err
	}

	idxFile := moduleName + ".cbuild-idx.yml"
	_, err = runner.ExecuteCommand(cbuild2cmakeBin, false, idxFile, "--zephyr")
	if err != nil {
		log.Error(err)
		return err
	}

	sourceModuleDir := filepath.Join(tmpDir, moduleName)
	if _, statErr := os.Stat(sourceModuleDir); statErr != nil {
		if os.IsNotExist(statErr) {
			err = errutils.New(errutils.ErrPathNotExist, sourceModuleDir)
		} else {
			err = statErr
		}
		log.Error(err)
		return err
	}

	destinationModuleDir := filepath.Join(originalDir, moduleName)
	if _, existsErr := os.Stat(destinationModuleDir); existsErr == nil {
		log.Warn("overwriting existing destination:", destinationModuleDir)
	}

	if err = os.RemoveAll(destinationModuleDir); err != nil {
		log.Error(err)
		return err
	}

	if err = os.Rename(sourceModuleDir, destinationModuleDir); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

var ZephyrCmd = &cobra.Command{
	Use:   "zephyr [options]",
	Short: "Generate Zephyr module files from clayer input",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateZephyrModule(cmd, args)
	},
}

func init() {
	ZephyrCmd.DisableFlagsInUseLine = true
	ZephyrCmd.Flags().StringP("module", "", "", "Module name")
	ZephyrCmd.Flags().StringSliceP("clayer", "", []string{}, "Comma-separated clayer files")
	ZephyrCmd.Flags().StringSliceP("packs", "", []string{defaultPack}, "Comma-separated pack identifiers")
	ZephyrCmd.Flags().StringP("toolchain", "", defaultToolchain, "Toolchain to be used")
	ZephyrCmd.Flags().StringP("device", "", defaultDevice, "Device to be used")
}
