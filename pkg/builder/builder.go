/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package builder

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	utils "cbuild/pkg/utils"

	log "github.com/sirupsen/logrus"
)

type BuilderInterface interface {
	Build() error
}

type Builder struct {
	Runner   utils.RunnerInterface
	CprjFile string
	Options  Options
}

type Options struct {
	IntDir    string
	OutDir    string
	LockFile  string
	LogFile   string
	Generator string
	Target    string
	Jobs      int
	Quiet     bool
	Debug     bool
	Clean     bool
	Schema    bool
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

func (b Builder) configLog() {
	log.SetLevel(log.InfoLevel)
	if b.Options.Debug {
		log.SetLevel(log.DebugLevel)
	} else if b.Options.Quiet {
		log.SetLevel(log.ErrorLevel)
	}
	if b.Options.LogFile != "" {
		logFile, err := os.Create(b.Options.LogFile)
		if err != nil {
			log.Warn("error creating log file")
		}
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(multiWriter)
	}
}

func (b Builder) checkCprj() error {
	if filepath.Ext(b.CprjFile) != ".cprj" {
		err := errors.New("missing required argument <project>.cprj")
		log.Error(err)
		return err
	} else {
		if _, err := os.Stat(b.CprjFile); os.IsNotExist(err) {
			log.Error("project file " + b.CprjFile + " does not exist")
			return err
		}
	}
	return nil
}

func (b Builder) getDirs() (dirs BuildDirs, err error) {
	if b.Options.IntDir != "" {
		dirs.intDir = b.Options.IntDir
	}
	if b.Options.OutDir != "" {
		dirs.outDir = b.Options.OutDir
	}
	intDir, outDir, err := GetCprjDirs(b.CprjFile)
	if err != nil {
		log.Error("error parsing file: " + b.CprjFile)
		return dirs, err
	}
	if dirs.intDir == "" {
		dirs.intDir = intDir
		if dirs.intDir == "" {
			dirs.intDir = "IntDir"
		}
		if !filepath.IsAbs(dirs.intDir) {
			dirs.intDir = filepath.Join(filepath.Dir(b.CprjFile), dirs.intDir)
		}
	}
	if dirs.outDir == "" {
		dirs.outDir = outDir
		if dirs.outDir == "" {
			dirs.outDir = "OutDir"
		}
		if !filepath.IsAbs(dirs.outDir) {
			dirs.outDir = filepath.Join(filepath.Dir(b.CprjFile), dirs.outDir)
		}
	}

	dirs.intDir, _ = filepath.Abs(dirs.intDir)
	dirs.outDir, _ = filepath.Abs(dirs.outDir)

	log.Debug("dirs.intDir: " + dirs.intDir)
	log.Debug("dirs.outDir: " + dirs.outDir)

	return dirs, err
}

func (b Builder) clean(dirs BuildDirs, vars InternalVars) (err error) {
	if _, err := os.Stat(dirs.intDir); !os.IsNotExist(err) {
		err = b.Runner.ExecuteCommand(vars.cbuildgenBin, false, "rmdir", dirs.intDir)
		if err != nil {
			log.Error("error executing 'cbuildgen rmdir'")
			return err
		}
	}
	if _, err := os.Stat(dirs.outDir); !os.IsNotExist(err) {
		err = b.Runner.ExecuteCommand(vars.cbuildgenBin, false, "rmdir", dirs.outDir)
		if err != nil {
			log.Error("error executing 'cbuildgen rmdir'")
			return err
		}
	}
	log.Info("finished successfully!")
	return nil
}

func (b Builder) getInternalVars() (vars InternalVars, err error) {

	vars.cprjPath = filepath.Dir(b.CprjFile)
	vars.cprjFilename = filepath.Base(b.CprjFile)
	vars.cprjFilename = strings.TrimSuffix(vars.cprjFilename, filepath.Ext(vars.cprjFilename))

	vars.binPath = os.Getenv("CMSIS_BUILD_ROOT")
	if vars.binPath == "" {
		vars.binPath, err = utils.GetExecutablePath()
		if err != nil {
			log.Error("executable path was not found")
			return vars, err
		}
	}
	if vars.binPath != "" {
		vars.binPath, _ = filepath.Abs(vars.binPath)
	}

	vars.etcPath = filepath.Clean(vars.binPath + "/../etc")
	if _, err := os.Stat(vars.etcPath); os.IsNotExist(err) {
		log.Error("etc directory was not found")
		return vars, err
	}
	if vars.etcPath != "" {
		vars.etcPath, _ = filepath.Abs(vars.etcPath)
	}

	var binExtension string
	if runtime.GOOS == "windows" {
		binExtension = ".exe"
	}

	vars.cbuildgenBin = filepath.Join(vars.binPath, "cbuildgen"+binExtension)
	if _, err := os.Stat(vars.cbuildgenBin); os.IsNotExist(err) {
		log.Error("cbuildgen was not found")
		return vars, err
	}

	cpackgetBin := filepath.Join(vars.binPath, "cpackget"+binExtension)
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

func (b Builder) getJobs() (jobs int) {
	jobs = runtime.NumCPU()
	if b.Options.Jobs > 0 {
		jobs = b.Options.Jobs
	}
	return jobs
}

func (b Builder) Build() error {

	b.configLog()

	b.CprjFile, _ = filepath.Abs(b.CprjFile)
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

	if b.Options.Clean {
		return b.clean(dirs, vars)
	}

	if b.Options.Schema {
		if vars.xmllintBin == "" {
			log.Warn("xmllint was not found, proceed without xml validation")
		} else {
			err = b.Runner.ExecuteCommand(vars.xmllintBin, b.Options.Quiet, "--schema", filepath.Join(vars.etcPath, "CPRJ.xsd"), b.CprjFile, "--noout")
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
	args = []string{"packlist", b.CprjFile, "--outdir=" + dirs.outDir, "--intdir=" + dirs.intDir}
	if b.Options.Quiet {
		args = append(args, "--quiet")
	}
	err = b.Runner.ExecuteCommand(vars.cbuildgenBin, b.Options.Quiet, args...)
	if err != nil {
		log.Error("error executing 'cbuildgen packlist'")
		return err
	}

	if _, err := os.Stat(vars.packlistFile); !os.IsNotExist(err) {
		if vars.cpackgetBin == "" {
			log.Error("cpackget was not found, missing packs cannot be downloaded")
			return err
		}
		args = []string{"pack", "add", "--agree-embedded-license", "--packs-list-filename", vars.packlistFile}
		if b.Options.Debug {
			args = append(args, "--verbose")
		} else if b.Options.Quiet {
			args = append(args, "--quiet")
		}

		err = b.Runner.ExecuteCommand(vars.cpackgetBin, b.Options.Quiet, args...)
		if err != nil {
			log.Error("error executing 'cpackget pack add'")
			return err
		}
	}

	args = []string{"cmake", b.CprjFile, "--outdir=" + dirs.outDir, "--intdir=" + dirs.intDir}
	if b.Options.Quiet {
		args = append(args, "--quiet")
	}
	if b.Options.LockFile != "" {
		lockFile, _ := filepath.Abs(b.Options.LockFile)
		args = append(args, "--update="+lockFile)
	}
	err = b.Runner.ExecuteCommand(vars.cbuildgenBin, b.Options.Quiet, args...)
	if err != nil {
		log.Error("error executing 'cbuildgen cmake'")
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
	err = b.Runner.ExecuteCommand(vars.cmakeBin, b.Options.Quiet, args...)
	if err != nil {
		log.Error("error executing 'cmake' configuration")
		return err
	}

	args = []string{"--build", dirs.intDir, "-j", fmt.Sprintf("%d", b.getJobs())}
	if b.Options.Target != "" {
		args = append(args, "--target", b.Options.Target)
	}
	err = b.Runner.ExecuteCommand(vars.cmakeBin, false, args...)
	if err != nil {
		log.Error("error executing 'cmake' build")
		return err
	}

	log.Info("build finished successfully!")
	return nil
}
