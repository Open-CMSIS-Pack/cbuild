/*
 * Copyright (c) 2022-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cproject

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	builder "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/builder"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
	utils "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type CprjBuilder struct {
	builder.BuilderParams
}

func (b CprjBuilder) checkCprj() error {
	fileExtension := filepath.Ext(b.InputFile)
	expectedExtension := ".cprj"
	if fileExtension != expectedExtension {
		return errutils.New(errutils.ErrInvalidFileExtension, fileExtension, expectedExtension)
	} else {
		if _, err := os.Stat(b.InputFile); os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (b CprjBuilder) clean(dirs builder.BuildDirs, vars builder.InternalVars) (err error) {
	fileName := filepath.Base(b.InputFile)
	fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))]
	log.Info("Cleaning context: \"" + fileName + "\"")

	if _, err := os.Stat(dirs.IntDir); !os.IsNotExist(err) {
		_, err = b.Runner.ExecuteCommand(vars.CbuildgenBin, false, "rmdir", dirs.IntDir)
		if err != nil {
			return err
		}
	}
	if _, err := os.Stat(dirs.OutDir); !os.IsNotExist(err) {
		_, err = b.Runner.ExecuteCommand(vars.CbuildgenBin, false, "rmdir", dirs.OutDir)
		if err != nil {
			return err
		}
	}
	log.Info("clean finished successfully!")
	return nil
}

func (b CprjBuilder) getDirs() (dirs builder.BuildDirs, err error) {
	if b.Options.IntDir != "" {
		dirs.IntDir = b.Options.IntDir
	}
	if b.Options.OutDir != "" {
		dirs.OutDir = b.Options.OutDir
	}

	if b.Options.Output != "" {
		dirs.IntDir = ""
		dirs.OutDir = ""
	}

	intDir, outDir, err := GetCprjDirs(b.InputFile)
	if err != nil {
		return dirs, err
	}
	if dirs.IntDir == "" {
		dirs.IntDir = intDir
		if dirs.IntDir == "" {
			dirs.IntDir = "IntDir"
		}
		if !filepath.IsAbs(dirs.IntDir) {
			dirs.IntDir = filepath.Join(filepath.Dir(b.InputFile), dirs.IntDir)
		}
	}
	if dirs.OutDir == "" {
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

func (b CprjBuilder) build() error {
	b.InputFile, _ = filepath.Abs(b.InputFile)
	b.InputFile = utils.NormalizePath(b.InputFile)

	err := b.checkCprj()
	if err != nil {
		return err
	}

	dirs, err := b.getDirs()
	if err != nil {
		return err
	}

	vars, err := b.GetInternalVars()
	if err != nil {
		return err
	}

	_ = utils.UpdateEnvVars(vars.BinPath, vars.EtcPath)

	if b.Options.Rebuild {
		err = b.clean(dirs, vars)
		if err != nil {
			return err
		}
	} else if b.Options.Clean {
		err = b.clean(dirs, vars)
		return err
	}

	if b.Options.Schema {
		if vars.XmllintBin == "" {
			log.Warn("xmllint was not found, proceed without xml validation")
		} else {
			_, err = b.Runner.ExecuteCommand(vars.XmllintBin, b.Options.Quiet, "--schema", filepath.Join(vars.EtcPath, "CPRJ.xsd"), b.InputFile, "--noout")
			if err != nil {
				return err
			}
		}
	}

	cprjFilename := filepath.Base(b.InputFile)
	cprjFilename = strings.TrimSuffix(cprjFilename, filepath.Ext(cprjFilename))
	packlistFile := filepath.Join(dirs.IntDir, cprjFilename+".cpinstall")
	log.Debug("vars.packlistFile: " + packlistFile)
	_ = os.Remove(packlistFile)
	_ = os.MkdirAll(dirs.IntDir, 0755)

	var args []string
	args = []string{"packlist", b.InputFile, "--outdir=" + dirs.OutDir, "--intdir=" + dirs.IntDir}
	if b.Options.Quiet {
		args = append(args, "--quiet")
	}
	if b.Options.UpdateRte {
		args = append(args, "--update-rte")
	}
	_, err = b.Runner.ExecuteCommand(vars.CbuildgenBin, false, args...)
	if err != nil {
		return err
	}

	if _, err := os.Stat(packlistFile); !os.IsNotExist(err) {
		if b.Options.Packs {
			if vars.CpackgetBin == "" {
				err = errutils.New(errutils.ErrBinaryNotFound, "cpackget", "missing packs cannot be downloaded")
				return err
			}
			args = []string{"add", "--agree-embedded-license", "--no-dependencies", "--packs-list-filename", packlistFile}
			if b.Options.Debug {
				args = append(args, "--verbose")
			} else if b.Options.Quiet {
				args = append(args, "--quiet")
			}
			_, err = b.Runner.ExecuteCommand(vars.CpackgetBin, b.Options.Quiet, args...)
			if err != nil {
				return err
			}
		} else {
			err = errutils.New(errutils.ErrMissingPacks)
			return err
		}
	}

	args = []string{"cmake", b.InputFile, "--outdir=" + dirs.OutDir, "--intdir=" + dirs.IntDir}
	if b.Options.Quiet {
		args = append(args, "--quiet")
	}
	if b.Options.LockFile != "" {
		lockFile, _ := filepath.Abs(b.Options.LockFile)
		args = append(args, "--update="+lockFile)
	}

	if b.Options.Debug {
		log.Debug("cbuildgen command: " + vars.CbuildgenBin + " " + strings.Join(args, " "))
	}

	_, err = b.Runner.ExecuteCommand(vars.CbuildgenBin, false, args...)
	if err != nil {
		return err
	}

	if _, err := os.Stat(dirs.IntDir + "/CMakeLists.txt"); errors.Is(err, os.ErrNotExist) {
		return err
	}

	if vars.CmakeBin == "" {
		err = errutils.New(errutils.ErrBinaryNotFound, "cmake", "")
		return err
	}

	if b.Options.Generator == "" {
		b.Options.Generator = "Ninja"
		if vars.NinjaBin == "" {
			err = errutils.New(errutils.ErrBinaryNotFound, "ninja", "")
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
		return err
	}

	args = []string{"--build", dirs.IntDir, "-j", fmt.Sprintf("%d", b.GetJobs())}

	if b.Setup {
		args = append(args, "--target", "database")
	}

	if b.Options.Target != "" {
		args = append(args, "--target", b.Options.Target)
	}
	if b.Options.Debug || b.Options.Verbose {
		args = append(args, "--verbose")
	}

	if b.Options.Debug {
		log.Debug("cmake build command: " + vars.CmakeBin + " " + strings.Join(args, " "))
	}

	_, err = b.Runner.ExecuteCommand(vars.CmakeBin, false, args...)
	if err != nil {
		return err
	}

	operation := "build"
	if b.Setup {
		operation = "setup"
	}
	log.Info(operation + " finished successfully!")
	return nil
}

func (b CprjBuilder) Build() (err error) {
	if err = b.build(); err != nil {
		log.Error(err)
	}
	return err
}
