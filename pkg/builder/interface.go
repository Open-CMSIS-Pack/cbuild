/*
 * Copyright (c) 2023-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package builder

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type BuilderParams struct {
	Runner         utils.RunnerInterface
	Options        Options
	InputFile      string
	InstallConfigs utils.Configurations
	Setup          bool
}

type Options struct {
	IntDir          string
	OutDir          string
	LockFile        string
	LogFile         string
	Generator       string
	Target          string
	Contexts        []string
	Filter          string
	Load            string
	Output          string
	Toolchain       string
	Jobs            int
	Quiet           bool
	Debug           bool
	Verbose         bool
	Clean           bool
	Schema          bool
	Packs           bool
	Rebuild         bool
	UpdateRte       bool
	UseContextSet   bool
	FrozenPacks     bool
	UseCbuild2CMake bool
}

type InternalVars struct {
	BinPath         string
	EtcPath         string
	CbuildgenBin    string
	Cbuild2cmakeBin string
	XmllintBin      string
	CpackgetBin     string
	CmakeBin        string
	NinjaBin        string
}

type BuildDirs struct {
	IntDir string
	OutDir string
}

func (b BuilderParams) GetInternalVars() (vars InternalVars, err error) {
	vars.BinPath = b.InstallConfigs.BinPath
	vars.EtcPath = b.InstallConfigs.EtcPath

	vars.CbuildgenBin = filepath.Join(vars.BinPath, "cbuildgen"+b.InstallConfigs.BinExtn)
	if _, err := os.Stat(vars.CbuildgenBin); os.IsNotExist(err) {
		log.Error("cbuildgen was not found")
		return vars, err
	}

	vars.Cbuild2cmakeBin = filepath.Join(vars.BinPath, "cbuild2cmake"+b.InstallConfigs.BinExtn)
	if _, err := os.Stat(vars.Cbuild2cmakeBin); os.IsNotExist(err) {
		log.Error("cbuild2cmake was not found")
		return vars, err
	}

	cpackgetBin := filepath.Join(vars.BinPath, "cpackget"+b.InstallConfigs.BinExtn)
	if _, err := os.Stat(cpackgetBin); !os.IsNotExist(err) {
		vars.CpackgetBin = cpackgetBin
	}

	vars.XmllintBin, _ = exec.LookPath("xmllint")
	vars.CmakeBin, _ = exec.LookPath("cmake")
	vars.NinjaBin, _ = exec.LookPath("ninja")

	log.Debug("vars.binPath: " + vars.BinPath)
	log.Debug("vars.etcPath: " + vars.EtcPath)
	log.Debug("vars.cbuildgenBin: " + vars.CbuildgenBin)
	log.Debug("vars.cpackgetBin: " + vars.CpackgetBin)
	log.Debug("vars.xmllintBin: " + vars.XmllintBin)
	log.Debug("vars.cmakeBin: " + vars.CmakeBin)
	log.Debug("vars.ninjaBin: " + vars.NinjaBin)

	return vars, err
}

func (b BuilderParams) GetJobs() (jobs int) {
	jobs = runtime.NumCPU()
	if b.Options.Jobs > 0 {
		jobs = b.Options.Jobs
	}
	return jobs
}

type IBuilderInterface interface {
	Build() error
}
