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
	"strings"

	builder "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	utils "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type CbuildIdxBuilder struct {
	builder.BuilderParams
}

func (b CbuildIdxBuilder) checkCbuildIdx() error {
	fileName := filepath.Base(b.InputFile)
	if !strings.HasSuffix(fileName, ".cbuild-idx.yml") {
		err := errors.New(".cbuild-idx.yml file not found")
		return err
	} else {
		if _, err := os.Stat(b.InputFile); os.IsNotExist(err) {
			log.Error("cbuild-idx file " + b.InputFile + " does not exist")
			return err
		}
	}
	return nil
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
		log.Error("file " + b.InputFile + " does not exist")
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
			log.Error("error parsing file: " + cbuildFile)
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

func (b CbuildIdxBuilder) Build() error {
	b.InputFile, _ = filepath.Abs(b.InputFile)
	b.InputFile = utils.NormalizePath(b.InputFile)
	err := b.checkCbuildIdx()
	if err != nil {
		return err
	}

	vars, err := b.GetInternalVars()
	if err != nil {
		return err
	}

	_ = utils.UpdateEnvVars(vars.BinPath, vars.EtcPath)

	if len(b.Options.Contexts) == 0 && b.BuildContext == "" {
		err = errors.New("error no context(s) to process")
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

	if vars.CmakeBin == "" {
		log.Error("cmake was not found")
		return err
	}
	if b.Options.Generator == "" {
		b.Options.Generator = "Ninja"
		if vars.NinjaBin == "" {
			log.Error("ninja was not found")
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
	} else if b.BuildContext != "" {
		args = append(args, "--target", b.BuildContext)
	}

	if b.Setup {
		args = append(args, "--target", b.BuildContext+"-database")
	}

	if b.Options.Generator == "Ninja" && !(b.Options.Debug || b.Options.Verbose) {
		args = append(args, "--", "--quiet")
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
