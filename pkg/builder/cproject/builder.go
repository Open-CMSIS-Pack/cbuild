/*
 * Copyright (c) 2022-2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cproject

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	builder "cbuild/pkg/builder"
	utils "cbuild/pkg/utils"

	log "github.com/sirupsen/logrus"
)

type CprjBuilder struct {
	builder.BuilderParams
}

type BuildDirs struct {
	intDir string
	outDir string
}

type InternalVars struct {
	cprjPath     string
	cprjFilename string
	packlistFile string
	binPath      string
	etcPath      string
	cbuildgenBin string
	xmllintBin   string
	cpackgetBin  string
	cmakeBin     string
	ninjaBin     string
}

func (b CprjBuilder) checkCprj() error {
	if filepath.Ext(b.InputFile) != ".cprj" {
		err := errors.New("missing required argument <project>.cprj")
		return err
	} else {
		if _, err := os.Stat(b.InputFile); os.IsNotExist(err) {
			log.Error("project file " + b.InputFile + " does not exist")
			return err
		}
	}
	return nil
}

func (b CprjBuilder) getDirs() (dirs BuildDirs, err error) {
	if b.Options.IntDir != "" {
		dirs.intDir = b.Options.IntDir
	}
	if b.Options.OutDir != "" {
		dirs.outDir = b.Options.OutDir
	}

	if b.Options.Output != "" {
		dirs.intDir = ""
		dirs.outDir = ""
	}

	intDir, outDir, err := GetCprjDirs(b.InputFile)
	if err != nil {
		log.Error("error parsing file: " + b.InputFile)
		return dirs, err
	}
	if dirs.intDir == "" {
		dirs.intDir = intDir
		if dirs.intDir == "" {
			dirs.intDir = "IntDir"
		}
		if !filepath.IsAbs(dirs.intDir) {
			dirs.intDir = filepath.Join(filepath.Dir(b.InputFile), dirs.intDir)
		}
	}
	if dirs.outDir == "" {
		dirs.outDir = outDir
		if dirs.outDir == "" {
			dirs.outDir = "OutDir"
		}
		if !filepath.IsAbs(dirs.outDir) {
			dirs.outDir = filepath.Join(filepath.Dir(b.InputFile), dirs.outDir)
		}
	}

	dirs.intDir, _ = filepath.Abs(dirs.intDir)
	dirs.outDir, _ = filepath.Abs(dirs.outDir)

	log.Debug("dirs.intDir: " + dirs.intDir)
	log.Debug("dirs.outDir: " + dirs.outDir)

	return dirs, err
}

func (b CprjBuilder) clean(dirs BuildDirs, vars InternalVars) (err error) {
	if _, err := os.Stat(dirs.intDir); !os.IsNotExist(err) {
		_, err = b.Runner.ExecuteCommand(vars.cbuildgenBin, false, "rmdir", dirs.intDir)
		if err != nil {
			log.Error("error executing 'cbuildgen rmdir'")
			return err
		}
	}
	if _, err := os.Stat(dirs.outDir); !os.IsNotExist(err) {
		_, err = b.Runner.ExecuteCommand(vars.cbuildgenBin, false, "rmdir", dirs.outDir)
		if err != nil {
			log.Error("error executing 'cbuildgen rmdir'")
			return err
		}
	}
	log.Info("clean finished successfully!")
	return nil
}

func (b CprjBuilder) getInternalVars() (vars InternalVars, err error) {

	vars.cprjPath = filepath.Dir(b.InputFile)
	vars.cprjFilename = filepath.Base(b.InputFile)
	vars.cprjFilename = strings.TrimSuffix(vars.cprjFilename, filepath.Ext(vars.cprjFilename))

	vars.binPath = b.InstallConfigs.BinPath
	vars.etcPath = b.InstallConfigs.EtcPath

	vars.cbuildgenBin = filepath.Join(vars.binPath, "cbuildgen"+b.InstallConfigs.BinExtn)
	if _, err := os.Stat(vars.cbuildgenBin); os.IsNotExist(err) {
		log.Error("cbuildgen was not found")
		return vars, err
	}

	cpackgetBin := filepath.Join(vars.binPath, "cpackget"+b.InstallConfigs.BinExtn)
	if _, err := os.Stat(cpackgetBin); !os.IsNotExist(err) {
		vars.cpackgetBin = cpackgetBin
	}

	vars.xmllintBin, _ = exec.LookPath("xmllint")
	vars.cmakeBin, _ = exec.LookPath("cmake")
	vars.ninjaBin, _ = exec.LookPath("ninja")

	log.Debug("vars.binPath: " + vars.binPath)
	log.Debug("vars.etcPath: " + vars.etcPath)
	log.Debug("vars.cbuildgenBin: " + vars.cbuildgenBin)
	log.Debug("vars.cpackgetBin: " + vars.cpackgetBin)
	log.Debug("vars.xmllintBin: " + vars.xmllintBin)
	log.Debug("vars.cmakeBin: " + vars.cmakeBin)
	log.Debug("vars.ninjaBin: " + vars.ninjaBin)

	return vars, err
}

func (b CprjBuilder) getJobs() (jobs int) {
	jobs = runtime.NumCPU()
	if b.Options.Jobs > 0 {
		jobs = b.Options.Jobs
	}
	return jobs
}

func (b CprjBuilder) Build() error {

	b.InputFile, _ = filepath.Abs(b.InputFile)
	err := b.checkCprj()
	if err != nil {
		return err
	}

	dirs, err := b.getDirs()
	if err != nil {
		return err
	}

	vars, err := b.getInternalVars()
	if err != nil {
		return err
	}

	_ = utils.UpdateEnvVars(vars.binPath, vars.etcPath)

	if b.Options.Rebuild {
		err = b.clean(dirs, vars)
		if err != nil {
			return err
		}
	} else if b.Options.Clean {
		return b.clean(dirs, vars)
	}

	if b.Options.Schema {
		if vars.xmllintBin == "" {
			log.Warn("xmllint was not found, proceed without xml validation")
		} else {
			_, err = b.Runner.ExecuteCommand(vars.xmllintBin, b.Options.Quiet, "--schema", filepath.Join(vars.etcPath, "CPRJ.xsd"), b.InputFile, "--noout")
			if err != nil {
				log.Error("error executing 'xmllint'")
				return err
			}
		}
	}

	vars.packlistFile = filepath.Join(dirs.intDir, vars.cprjFilename+".cpinstall")
	log.Debug("vars.packlistFile: " + vars.packlistFile)
	_ = os.Remove(vars.packlistFile)
	_ = os.MkdirAll(dirs.intDir, 0755)

	var args []string
	args = []string{"packlist", b.InputFile, "--outdir=" + dirs.outDir, "--intdir=" + dirs.intDir}
	if b.Options.Quiet {
		args = append(args, "--quiet")
	}
	if b.Options.UpdateRte {
		args = append(args, "--update-rte")
	}
	_, err = b.Runner.ExecuteCommand(vars.cbuildgenBin, false, args...)
	if err != nil {
		log.Error("error executing 'cbuildgen packlist'")
		return err
	}

	if _, err := os.Stat(vars.packlistFile); !os.IsNotExist(err) {
		if b.Options.Packs {
			if vars.cpackgetBin == "" {
				err := errors.New("cpackget was not found, missing packs cannot be downloaded")
				return err
			}
			args = []string{"add", "--agree-embedded-license", "--no-dependencies", "--packs-list-filename", vars.packlistFile}
			if b.Options.Debug {
				args = append(args, "--verbose")
			} else if b.Options.Quiet {
				args = append(args, "--quiet")
			}
			_, err = b.Runner.ExecuteCommand(vars.cpackgetBin, b.Options.Quiet, args...)
			if err != nil {
				log.Error("error executing 'cpackget add'")
				return err
			}
		} else {
			err := errors.New("missing packs must be installed, rerun cbuild with the --packs option")
			log.Error(err)
			return err
		}
	}

	args = []string{"cmake", b.InputFile, "--outdir=" + dirs.outDir, "--intdir=" + dirs.intDir}
	if b.Options.Quiet {
		args = append(args, "--quiet")
	}
	if b.Options.LockFile != "" {
		lockFile, _ := filepath.Abs(b.Options.LockFile)
		args = append(args, "--update="+lockFile)
	}

	if b.Options.Debug {
		log.Debug("cbuildgen command: " + vars.cbuildgenBin + " " + strings.Join(args, " "))
	}

	_, err = b.Runner.ExecuteCommand(vars.cbuildgenBin, false, args...)
	if err != nil {
		log.Error("error executing 'cbuildgen cmake'")
		return err
	}

	if _, err := os.Stat(dirs.intDir + "/CMakeLists.txt"); errors.Is(err, os.ErrNotExist) {
		return err
	}

	if vars.cmakeBin == "" {
		log.Error("cmake was not found")
		return err
	}

	if b.Options.Generator == "" {
		b.Options.Generator = "Ninja"
		if vars.ninjaBin == "" {
			log.Error("ninja was not found")
			return err
		}
	}

	args = []string{"-G", b.Options.Generator, "-S", dirs.intDir, "-B", dirs.intDir}
	if b.Options.Debug {
		args = append(args, "-Wdev")
	} else {
		args = append(args, "-Wno-dev")
	}

	if b.Options.Debug {
		log.Debug("cmake configuration command: " + vars.cmakeBin + " " + strings.Join(args, " "))
	}

	_, err = b.Runner.ExecuteCommand(vars.cmakeBin, b.Options.Quiet, args...)
	if err != nil {
		log.Error("error executing 'cmake' configuration")
		return err
	}

	args = []string{"--build", dirs.intDir, "-j", fmt.Sprintf("%d", b.getJobs())}
	if b.Options.Target != "" {
		args = append(args, "--target", b.Options.Target)
	}
	if b.Options.Debug || b.Options.Verbose {
		args = append(args, "--verbose")
	}

	if b.Options.Debug {
		log.Debug("cmake build command: " + vars.cmakeBin + " " + strings.Join(args, " "))
	}

	_, err = b.Runner.ExecuteCommand(vars.cmakeBin, false, args...)
	if err != nil {
		log.Error("error executing 'cmake' build")
		return err
	}

	log.Info("build finished successfully!")
	return nil
}
