/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package builder

import (
	"os"
	"os/exec"
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

func ExecuteCommand(program string, quiet bool, args ...string) error {
	cmd := exec.Command(program, args...)
	if !quiet {
		cmd.Stdout = log.StandardLogger().Out
		cmd.Stderr = log.StandardLogger().Out
	}
	err := cmd.Run()
	return err
}

func UpdateEnvVars(binPath string, etcPath string) (env EnvVars, err error) {
	env.PackRoot = os.Getenv("CMSIS_PACK_ROOT")
	if env.PackRoot == "" {
		env.PackRoot = GetDefaultCmsisPackRoot()
		if env.PackRoot != "" {
			os.Setenv("CMSIS_PACK_ROOT", env.PackRoot)
		}
	}
	env.CompilerRoot = os.Getenv("CMSIS_COMPILER_ROOT")
	if env.CompilerRoot == "" {
		env.CompilerRoot = etcPath
		os.Setenv("CMSIS_COMPILER_ROOT", env.CompilerRoot)
	}
	env.BuildRoot = os.Getenv("CMSIS_BUILD_ROOT")
	if env.BuildRoot == "" {
		env.BuildRoot = binPath
		os.Setenv("CMSIS_BUILD_ROOT", env.BuildRoot)
	}
	log.Debug("CMSIS_PACK_ROOT: " + env.PackRoot)
	log.Debug("CMSIS_COMPILER_ROOT: " + env.CompilerRoot)
	log.Debug("CMSIS_BUILD_ROOT: " + env.BuildRoot)
	return env, err
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
