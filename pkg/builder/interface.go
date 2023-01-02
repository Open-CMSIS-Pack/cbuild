/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package builder

import (
	"cbuild/pkg/utils"
)

type BuilderParams struct {
	Runner         utils.RunnerInterface
	Options        Options
	InputFile      string
	InstallConfigs utils.Configurations
}

type Options struct {
	IntDir    string
	OutDir    string
	LockFile  string
	LogFile   string
	Generator string
	Target    string
	Context   string
	Filter    string
	Load      string
	Jobs      int
	Quiet     bool
	Debug     bool
	Verbose   bool
	Clean     bool
	Schema    bool
	Packs     bool
	Rebuild   bool
	UpdateRte bool
}

type IBuilderInterface interface {
	Build() error
}

// func NewBuilder(runner utils.RunnerInterface, options Options,
// 	inputFile string) (bldr IBuilderInterface, err error) {

// 	configs, err := utils.GetInstallConfigs()
// 	if err != nil {
// 		return bldr, err
// 	}

// 	params := BuilderParams{
// 		Runner:         utils.Runner{},
// 		Options:        options,
// 		InputFile:      inputFile,
// 		InstallConfigs: configs,
// 	}

// 	fileExtension := filepath.Ext(inputFile)

// 	if fileExtension == ".cprj" {
// 		bldr = cproject.CPRJBuilder{BuilderParams: params}
// 	} else if fileExtension == ".yml" {
// 		bldr = csolution.CSolutionBuilder{BuilderParams: params}
// 	} else {
// 		err = errors.New("invalid file argument")
// 	}

// 	return bldr, err
// }
