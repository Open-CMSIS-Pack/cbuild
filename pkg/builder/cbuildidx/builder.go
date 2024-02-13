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
		if err != nil {
			log.Error("error executing 'cmake' clean for " + dir)
		}
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

func (b CbuildIdxBuilder) getDirs() (dirs builder.BuildDirs, err error) {
	if _, err := os.Stat(b.InputFile); os.IsNotExist(err) {
		log.Error("file " + b.InputFile + " does not exist")
		return dirs, err
	}

	if len(b.Options.Contexts) != 1 {
		err = errors.New("error invalid context(s) process request")
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
		path := filepath.Dir(b.InputFile)
		cbuildFile := filepath.Join(path, b.Options.Contexts[0]+".cbuild.yml")
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
			dirs.OutDir = filepath.Join(filepath.Dir(b.InputFile), dirs.OutDir)
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
	err := b.checkCbuildIdx()
	if err != nil {
		return err
	}

	vars, err := b.GetInternalVars()
	if err != nil {
		return err
	}

	_ = utils.UpdateEnvVars(vars.BinPath, vars.EtcPath)

	dirs, err := b.getDirs()
	if err != nil {
		return err
	}

	if b.Options.Rebuild {
		err = b.clean(dirs, vars)
		if err != nil {
			return err
		}
	} else if b.Options.Clean {
		return b.clean(dirs, vars)
	}

	args := []string{b.InputFile}
	if b.Options.Debug {
		log.Debug("cbuild2cmake command: " + vars.Cbuild2cmakeBin + " " + strings.Join(args, " "))
	}

	_, err = b.Runner.ExecuteCommand(vars.Cbuild2cmakeBin, false, args...)
	if err != nil {
		log.Error("error executing 'cbuild2cmake " + b.InputFile + "'")
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

	args = []string{"-G", b.Options.Generator, "-S", dirs.IntDir, "-B", dirs.IntDir}
	if b.Options.Debug {
		args = append(args, "-Wdev")
	} else {
		args = append(args, "-Wno-dev")
	}

	if b.Options.Debug {
		log.Debug("cmake configuration command: " + vars.CmakeBin + " " + strings.Join(args, " "))
	}

	_, err = b.Runner.ExecuteCommand(vars.CmakeBin, b.Options.Quiet, args...)
	if err != nil {
		log.Error("error executing 'cmake' configuration")
		return err
	}

	args = []string{"--build", dirs.IntDir, "-j", fmt.Sprintf("%d", b.GetJobs())}

	if len(b.Options.Contexts) == 1 {
		args = append(args, "--target", b.Options.Contexts[0])
		args = append(args, "--target", b.Options.Contexts[0]+"-database")
	} else {
		err = errors.New("error invalid context(s) process request")
		return err
	}
	if b.Options.Debug || b.Options.Verbose {
		args = append(args, "--verbose")
	}

	if b.Options.Debug {
		log.Debug("cmake build command: " + vars.CmakeBin + " " + strings.Join(args, " "))
	}

	_, err = b.Runner.ExecuteCommand(vars.CmakeBin, false, args...)
	if err != nil {
		log.Error("error executing 'cmake' build")
		return err
	}

	log.Info("build finished successfully!")
	return nil
}
