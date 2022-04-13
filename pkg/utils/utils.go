/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type EnvVars struct {
	PackRoot     string
	CompilerRoot string
	BuildRoot    string
}

func GetExecutablePath() (string, error) {
	exec, err := os.Executable()
	if err != nil {
		return "", err
	}
	execReal, err := filepath.EvalSymlinks(exec)
	if err != nil {
		return "", err
	}
	executablePath := filepath.Dir(execReal)
	return executablePath, nil
}

func UpdateEnvVars(binPath string, etcPath string) (env EnvVars) {
	env.PackRoot = os.Getenv("CMSIS_PACK_ROOT")
	if env.PackRoot == "" {
		packRoot := GetDefaultCmsisPackRoot()
		if packRoot != "" {
			env.PackRoot, _ = filepath.Abs(packRoot)
			os.Setenv("CMSIS_PACK_ROOT", env.PackRoot)
		}
	}
	env.CompilerRoot = os.Getenv("CMSIS_COMPILER_ROOT")
	if env.CompilerRoot == "" {
		env.CompilerRoot, _ = filepath.Abs(etcPath)
		os.Setenv("CMSIS_COMPILER_ROOT", env.CompilerRoot)
	}
	env.BuildRoot = os.Getenv("CMSIS_BUILD_ROOT")
	if env.BuildRoot == "" {
		env.BuildRoot, _ = filepath.Abs(binPath)
		os.Setenv("CMSIS_BUILD_ROOT", env.BuildRoot)
	}
	log.Debug("CMSIS_PACK_ROOT: " + env.PackRoot)
	log.Debug("CMSIS_COMPILER_ROOT: " + env.CompilerRoot)
	log.Debug("CMSIS_BUILD_ROOT: " + env.BuildRoot)
	return env
}

func GetDefaultCmsisPackRoot() (root string) {
	root = os.Getenv("LOCALAPPDATA")
	if root == "" {
		root = os.Getenv("XDG_CACHE_HOME")
	}
	if root == "" {
		root = os.Getenv("HOME")
	}
	if root != "" {
		root = filepath.Clean(root + "/Arm/Packs")
	}
	return root
}
