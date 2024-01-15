/*
 * Copyright (c) 2023-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package builder

import (
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
)

type BuilderParams struct {
	Runner         utils.RunnerInterface
	Options        Options
	InputFile      string
	InstallConfigs utils.Configurations
}

type Options struct {
	IntDir        string
	OutDir        string
	LockFile      string
	LogFile       string
	Generator     string
	Target        string
	Contexts      []string
	Filter        string
	Load          string
	Output        string
	Toolchain     string
	Jobs          int
	Quiet         bool
	Debug         bool
	Verbose       bool
	Clean         bool
	Schema        bool
	Packs         bool
	Rebuild       bool
	UpdateRte     bool
	UseContextSet bool
	FrozenPacks   bool
}

type IBuilderInterface interface {
	Build() error
}
