/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package builder

import (
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

type CmdOptions struct {
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

func configLog(cmdOptions CmdOptions) {
	log.SetLevel(log.InfoLevel)
	if cmdOptions.Quiet {
		log.SetLevel(log.ErrorLevel)
	} else if cmdOptions.Debug {
		log.SetLevel(log.DebugLevel)
	}
	if cmdOptions.LogFile != "" {
		logFile, err := os.Create(cmdOptions.LogFile)
		if err != nil {
			log.Warn("error creating log file")
		}
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(multiWriter)
	}
}

func checkCprj(cprjFile string, cmdOptions CmdOptions) error {
	if _, err := os.Stat(cprjFile); os.IsNotExist(err) {
		if filepath.Ext(cprjFile) != ".cprj" {
			log.Error("missing required argument <project>.cprj")
		} else {
			log.Error("project file " + cprjFile + " does not exist")
		}
		return err
	}
	return nil
}

func getDirs(cprjFile string, cmdOptions CmdOptions) (dirs BuildDirs, err error) {
	if cmdOptions.IntDir != "" {
		dirs.intDir = cmdOptions.IntDir
	}
	if cmdOptions.OutDir != "" {
		dirs.outDir = cmdOptions.OutDir
	}
	intDir, outDir, err := GetCprjDirs(cprjFile)
	if err != nil {
		log.Error("error parsing file: " + cprjFile)
		return dirs, err
	}
	if dirs.intDir == "" {
		dirs.intDir = intDir
		if dirs.intDir == "" {
			dirs.intDir = "IntDir"
		}
	}
	if dirs.outDir == "" {
		dirs.outDir = outDir
		if dirs.outDir == "" {
			dirs.outDir = "OutDir"
		}
	}
	dirs.intDir, _ = filepath.Abs(dirs.intDir)
	dirs.outDir, _ = filepath.Abs(dirs.outDir)

	log.Debug("dirs.intDir: " + dirs.intDir)
	log.Debug("dirs.outDir: " + dirs.outDir)

	return dirs, err
}

func clean(dirs BuildDirs, vars InternalVars) (err error) {
	if _, err := os.Stat(dirs.intDir); !os.IsNotExist(err) {
		err = utils.ExecuteCommand(vars.cbuildgenBin, false, "rmdir", dirs.intDir)
		if err != nil {
			log.Error("error executing 'cbuildgen rmdir'")
			return err
		}
	}
	if _, err := os.Stat(dirs.outDir); !os.IsNotExist(err) {
		err = utils.ExecuteCommand(vars.cbuildgenBin, false, "rmdir", dirs.outDir)
		if err != nil {
			log.Error("error executing 'cbuildgen rmdir'")
			return err
		}
	}
	log.Info("finished successfully!")
	return nil
}

func getInternalVars(cprjFile string, cmdOptions CmdOptions) (vars InternalVars, err error) {

	vars.cprjPath = filepath.Dir(cprjFile)
	vars.cprjFilename = filepath.Base(cprjFile)
	vars.cprjFilename = strings.TrimSuffix(vars.cprjFilename, filepath.Ext(vars.cprjFilename))

	vars.binPath, err = utils.GetExecutablePath()
	if err != nil {
		log.Error("executable path was not found")
		return vars, err
	}

	var binExtension string
	if runtime.GOOS == "windows" {
		binExtension = ".exe"
	}

	vars.cbuildgenBin, err = exec.LookPath("cbuildgen")
	if err != nil {
		vars.cbuildgenBin = filepath.Join(vars.binPath, "cbuildgen"+binExtension)
		if _, err := os.Stat(vars.cbuildgenBin); os.IsNotExist(err) {
			log.Error("cbuildgen was not found")
			return vars, err
		}
	}

	vars.etcPath = filepath.Clean(vars.binPath + "/../etc")
	if _, err := os.Stat(vars.etcPath); os.IsNotExist(err) {
		vars.etcPath = filepath.Clean(filepath.Dir(vars.cbuildgenBin) + "/../etc")
		if _, err := os.Stat(vars.cbuildgenBin); os.IsNotExist(err) {
			log.Error("etc directory was not found")
			return vars, err
		}
	}

	vars.cpackgetBin, err = exec.LookPath("cpackget")
	if err != nil {
		vars.cpackgetBin = filepath.Join(vars.binPath, "cpackget"+binExtension)
		if _, err := os.Stat(vars.cpackgetBin); os.IsNotExist(err) {
			log.Error("cpackget was not found")
			return vars, err
		}
	}

	vars.xmllintBin, _ = exec.LookPath("xmllint")
	vars.cmakeBin, _ = exec.LookPath("cmake")
	vars.ninjaBin, _ = exec.LookPath("ninja")

	log.Debug("vars.binPath: " + vars.binPath)
	log.Debug("vars.etcPath: " + vars.etcPath)
	log.Debug("vars.cbuildgenBin: " + vars.cbuildgenBin)
	log.Debug("vars.xmllintBin: " + vars.xmllintBin)
	log.Debug("vars.cpackgetBin: " + vars.cpackgetBin)
	log.Debug("vars.cmakeBin: " + vars.cmakeBin)
	log.Debug("vars.ninjaBin: " + vars.ninjaBin)

	_, err = utils.UpdateEnvVars(vars.binPath, vars.etcPath)
	return vars, err
}

func getJobs(cmdOptions CmdOptions) (jobs int) {
	jobs = runtime.NumCPU()
	if cmdOptions.Jobs != 0 {
		jobs = cmdOptions.Jobs
	}
	return jobs
}

func Build(cprjFile string, cmdOptions CmdOptions) error {

	configLog(cmdOptions)

	cprjFile, _ = filepath.Abs(cprjFile)
	err := checkCprj(cprjFile, cmdOptions)
	if err != nil {
		return err
	}

	dirs, err := getDirs(cprjFile, cmdOptions)
	if err != nil {
		return err
	}

	vars, err := getInternalVars(cprjFile, cmdOptions)
	if err != nil {
		return err
	}

	if cmdOptions.Clean {
		return clean(dirs, vars)
	}

	if vars.xmllintBin == "" {
		log.Warn("xmllint was not found, proceed without xml validation")
	} else {
		err = utils.ExecuteCommand(vars.xmllintBin, cmdOptions.Quiet, "--schema", filepath.Join(vars.etcPath, "CPRJ.xsd"), cprjFile, "--noout")
		if err != nil {
			log.Error("error executing 'xmllint'")
			return err
		}
	}

	vars.packlistFile = filepath.Join(dirs.intDir, vars.cprjFilename+".cpinstall")
	log.Debug("vars.packlistFile: " + vars.packlistFile)
	_ = os.Remove(vars.packlistFile)

	var args []string
	args = []string{"packlist", cprjFile, "--outdir=" + dirs.outDir, "--intdir=" + dirs.intDir}
	if cmdOptions.Quiet {
		args = append(args, "--quiet")
	}
	err = utils.ExecuteCommand(vars.cbuildgenBin, cmdOptions.Quiet, args...)
	if err != nil {
		log.Error("error executing 'cbuildgen packlist'")
		return err
	}

	if _, err := os.Stat(vars.packlistFile); !os.IsNotExist(err) {
		if vars.cpackgetBin == "" {
			log.Error("cpackget was not found, missing packs cannot be downloaded")
			return err
		}
		args = []string{"pack", "add", "-v", "--agree-embedded-license", "--packs-list-filename", vars.packlistFile}
		if cmdOptions.Quiet {
			args = append(args, "--quiet")
		}
		err = utils.ExecuteCommand(vars.cpackgetBin, cmdOptions.Quiet, args...)
		if err != nil {
			log.Error("error executing 'cpackget pack add'")
			return err
		}
	}

	args = []string{"cmake", cprjFile, "--outdir=" + dirs.outDir, "--intdir=" + dirs.intDir}
	if cmdOptions.Quiet {
		args = append(args, "--quiet")
	}
	if cmdOptions.LockFile != "" {
		lockFile, _ := filepath.Abs(cmdOptions.LockFile)
		args = append(args, "--update="+lockFile)
	}
	err = utils.ExecuteCommand(vars.cbuildgenBin, cmdOptions.Quiet, args...)
	if err != nil {
		log.Error("error executing 'cbuildgen cmake'")
		return err
	}

	if vars.cmakeBin == "" {
		log.Error("cmake was not found")
		return err
	}

	if cmdOptions.Generator == "Ninja" && vars.ninjaBin == "" {
		log.Error("ninja was not found")
		return err
	}

	args = []string{"-G", cmdOptions.Generator, "-S", dirs.intDir, "-B", dirs.intDir}
	if cmdOptions.Quiet {
		args = append(args, "-Wno-dev")
	} else {
		args = append(args, "-Wdev")
	}
	err = utils.ExecuteCommand(vars.cmakeBin, cmdOptions.Quiet, args...)
	if err != nil {
		log.Error("error executing 'cmake' configuration")
		return err
	}

	args = []string{"--build", dirs.intDir, "-j", fmt.Sprintf("%d", getJobs(cmdOptions))}
	if cmdOptions.Target != "" {
		args = append(args, "--target", cmdOptions.Target)
	}
	err = utils.ExecuteCommand(vars.cmakeBin, false, args...)
	if err != nil {
		log.Error("error executing 'cmake' build")
		return err
	}

	log.Info("build finished successfully!")
	return nil
}
