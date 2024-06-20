/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cbuildidx

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	builder "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
	utils "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
)

const NinjaVersion = "1.11.1"

type CbuildIdxBuilder struct {
	builder.BuilderParams
}

func (b CbuildIdxBuilder) clean(dirs builder.BuildDirs, vars builder.InternalVars) (err error) {
	removeDirectory := func(dir string) error {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return nil
		}
		args := []string{"-E", "remove_directory", dir}
		_, err = b.Runner.ExecuteCommand(vars.CmakeBin, false, args...)
		return err
	}

	if err := removeDirectory(dirs.IntDir); err != nil {
		return err
	}

	if err := removeDirectory(dirs.OutDir); err != nil {
		return err
	}

	log.Info("clean finished successfully!")
	return nil
}

func (b CbuildIdxBuilder) getDirs(context string) (dirs builder.BuildDirs, err error) {
	if _, err := os.Stat(b.InputFile); os.IsNotExist(err) {
		return dirs, err
	}

	if b.Options.OutDir != "" {
		dirs.OutDir = b.Options.OutDir
	}

	if b.Options.Output != "" {
		dirs.IntDir = ""
		dirs.OutDir = ""
	}

	// cbuild2cmake generates cmake files under fixed tmp directory
	dirs.IntDir = "tmp"
	dirs.IntDir = filepath.Join(filepath.Dir(b.InputFile), dirs.IntDir)

	if dirs.OutDir == "" {
		// get output directory from cbuild.yml file
		data, err := utils.ParseCbuildIndexFile(b.InputFile)
		if err != nil {
			return dirs, err
		}
		var cbuildFile string
		for _, cbuild := range data.BuildIdx.Cbuilds {
			if context == cbuild.Project+cbuild.Configuration {
				cbuildFile = cbuild.Cbuild
				break
			}
		}
		path := filepath.Dir(b.InputFile)
		cbuildFile = filepath.Join(path, cbuildFile)
		_, outDir, err := GetBuildDirs(cbuildFile)
		if err != nil {
			return dirs, err
		}

		dirs.OutDir = outDir
		if dirs.OutDir == "" {
			dirs.OutDir = "OutDir"
		}
		if !filepath.IsAbs(dirs.OutDir) {
			dirs.OutDir = filepath.Join(filepath.Dir(cbuildFile), dirs.OutDir)
		}
	}

	dirs.IntDir, _ = filepath.Abs(dirs.IntDir)
	dirs.OutDir, _ = filepath.Abs(dirs.OutDir)

	log.Debug("dirs.IntDir: " + dirs.IntDir)
	log.Debug("dirs.OutDir: " + dirs.OutDir)

	return dirs, err
}

func (b CbuildIdxBuilder) build() error {
	b.InputFile, _ = filepath.Abs(b.InputFile)
	b.InputFile = utils.NormalizePath(b.InputFile)

	_, err := utils.FileExists(b.InputFile)
	if err != nil {
		return err
	}

	vars, err := b.GetInternalVars()
	if err != nil {
		return err
	}

	_ = utils.UpdateEnvVars(vars.BinPath, vars.EtcPath)

	if len(b.Options.Contexts) == 0 && b.BuildContext == "" {
		err = errutils.New(errutils.ErrNoContextFound)
		return err
	}

	dirs := builder.BuildDirs{
		IntDir: filepath.Join(filepath.Dir(b.InputFile), "tmp"),
	}

	if b.Options.Clean {
		dirs, err := b.getDirs(b.BuildContext)
		if err != nil {
			return err
		}

		log.Info("Cleaning context: \"" + b.BuildContext + "\"")
		if err := b.clean(dirs, vars); err != nil {
			return err
		}
		return nil
	}

	if vars.CmakeBin == "" {
		err = errutils.New(errutils.ErrBinaryNotFound, "cmake", "")
		return err
	}

	args := []string{b.InputFile}
	if b.Options.UseContextSet {
		args = append(args, "--context-set")
	}
	if b.Options.Debug {
		args = append(args, "--debug")
		log.Debug("cbuild2cmake command: " + vars.Cbuild2cmakeBin + " " + strings.Join(args, " "))
	}

	_, err = b.Runner.ExecuteCommand(vars.Cbuild2cmakeBin, !(b.Options.Debug || b.Options.Verbose), args...)
	if err != nil {
		return err
	}
	if _, err := os.Stat(dirs.IntDir + "/CMakeLists.txt"); errors.Is(err, os.ErrNotExist) {
		return err
	}

	if b.Options.Generator == "" {
		b.Options.Generator = "Ninja"
		if vars.NinjaBin == "" {
			err = errutils.New(errutils.ErrBinaryNotFound, "ninja", "")
			return err
		}
	}

	// CMake configuration command
	args = []string{"-G", b.Options.Generator, "-S", dirs.IntDir, "-B", dirs.IntDir}
	if b.Options.Debug {
		args = append(args, "-Wdev")
	} else {
		args = append(args, "-Wno-dev")
	}

	if b.Options.Debug {
		log.Debug("cmake configuration command: " + vars.CmakeBin + " " + strings.Join(args, " "))
	}

	_, err = b.Runner.ExecuteCommand(vars.CmakeBin, !(b.Options.Debug || b.Options.Verbose), args...)
	if err != nil {
		return err
	}

	// CMake build target(s) command
	args = []string{"--build", dirs.IntDir, "-j", fmt.Sprintf("%d", b.GetJobs())}

	if b.Options.Target != "" {
		args = append(args, "--target", b.Options.Target)
	} else if b.Setup {
		args = append(args, "--target", b.BuildContext+"-database")
	} else if b.BuildContext != "" {
		args = append(args, "--target", b.BuildContext)
	}

	if b.Options.Generator == "Ninja" && !(b.Options.Debug || b.Options.Verbose) {
		isVersionGreaterorEqual, err := b.validateNinjaVersion(NinjaVersion)
		if err != nil {
			return err
		}

		if isVersionGreaterorEqual {
			args = append(args, "--", "--quiet")
		} else {
			log.Warn(errutils.WarnNinjaVersion)
		}
	}

	if b.Options.Debug {
		log.Debug("cmake build command: " + vars.CmakeBin + " " + strings.Join(args, " "))
	}

	_, err = b.Runner.ExecuteCommand(vars.CmakeBin, false, args...)
	if err != nil {
		return err
	}

	log.Info("build finished successfully!")
	return nil
}

func (b CbuildIdxBuilder) Build() (err error) {
	if err = b.build(); err != nil {
		log.Error(err)
	}
	return err
}

func (b CbuildIdxBuilder) validateNinjaVersion(refVersion string) (bool, error) {
	// Fetch installed version of ninja
	version, err := b.getNinjaVersion()
	if err != nil {
		return false, err
	}

	// Compare with fixed 1.11.1 version
	result, err := b.compareVersions(version, refVersion)
	if err != nil {
		return false, err
	}

	// Installed ninja version is lesser
	if result == -1 {
		return false, nil
	}

	// Installed version is greater or equal
	return true, nil
}

// Retrieves ninja version
func (b CbuildIdxBuilder) getNinjaVersion() (string, error) {
	versionStr, err := b.Runner.ExecuteCommand("ninja", true, "--version")
	if err != nil {
		return "", errutils.New(errutils.ErrBinaryNotFound, "ninja", "")
	}

	re := regexp.MustCompile(`^[\d]+.[\d+]+.[\d+]`)
	version := re.FindString(versionStr)
	if version == "" {
		return "", errutils.New(errutils.ErrNinjaVersionNotFound)
	}
	return version, nil
}

// Compare compares this version to another version. This
// returns -1, 0, or 1 if this version is smaller, equal,
// or larger than the other version, respectively
// or error when invalid input
func (b CbuildIdxBuilder) compareVersions(v1, v2 string) (int, error) {
	version1, err := version.NewSemver(v1)
	if err != nil {
		return 0, err
	}
	version2, err := version.NewSemver(v2)
	if err != nil {
		return 0, err
	}

	return version1.Compare(version2), nil
}
