/*
 * Copyright (c) 2023-2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package builder

import (
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/logger"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
)

type BuilderParams struct {
	Runner         utils.RunnerInterface
	Options        Options
	InputFile      string
	InstallConfigs utils.Configurations
	Setup          bool
	BuildContext   string
	ImageOnly      bool
	Executes       bool
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
	TargetSet       string
	Jobs            int
	Quiet           bool
	Debug           bool
	Verbose         bool
	Clean           bool
	SchemaChk       bool
	Packs           bool
	Rebuild         bool
	UpdateRte       bool
	UseContextSet   bool
	FrozenPacks     bool
	UseCbuild2CMake bool
	NoDatabase      bool
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
		return vars, err
	}

	vars.Cbuild2cmakeBin = filepath.Join(vars.BinPath, "cbuild2cmake"+b.InstallConfigs.BinExtn)
	if _, err := os.Stat(vars.Cbuild2cmakeBin); os.IsNotExist(err) {
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

type IBuilderInterface interface {
	Build() error
	Clean() error
}
