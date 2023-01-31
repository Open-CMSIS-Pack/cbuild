/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type EnvVars struct {
	PackRoot     string
	CompilerRoot string
	BuildRoot    string
}

type ContextItem struct {
	ProjectName string
	BuildType   string
	TargetType  string
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
	if runtime.GOOS == "windows" {
		root = os.Getenv("LOCALAPPDATA")
		if root == "" {
			root = os.Getenv("USERPROFILE")
			if root != "" {
				root = root + "\\AppData\\Local"
			}
		}
		if root != "" {
			root = root + "\\Arm\\Packs"
		}
	} else {
		root = os.Getenv("XDG_CACHE_HOME")
		if root == "" {
			root = os.Getenv("HOME")
			if root != "" {
				root = root + "/.cache"
			}
		}
		if root != "" {
			root = root + "/arm/packs"
		}
	}
	return filepath.Clean(root)
}

func ParseContext(context string) (item ContextItem, err error) {
	if context == "" {
		return ContextItem{}, errors.New("invalid context")
	}

	buildIdx := strings.Index(context, ".")
	targetIdx := strings.Index(context, "+")

	if buildIdx > targetIdx {
		err = errors.New("invalid context")
	} else {
		projectName := context[:buildIdx]
		buildType := context[buildIdx+1 : targetIdx]
		targetType := context[targetIdx+1:]
		if projectName == "" || buildType == "" || targetType == "" {
			err = errors.New("invalid context")
		} else {
			item.ProjectName = projectName
			item.BuildType = buildType
			item.TargetType = targetType
		}
	}
	return
}

type CbuildIndex struct {
	BuildIdx struct {
		GeneratedBy string `yaml:"generated-by"`
		Cdefault    string `yaml:"cdefault"`
		Csolution   string `yaml:"csolution"`
		Cprojects   []struct {
			Cproject string `yaml:"cproject"`
		} `yaml:"cprojects"`
		Licenses interface{} `yaml:"licenses"`
		Cbuilds  []struct {
			Cbuild string `yaml:"cbuild"`
		} `yaml:"cbuilds"`
	} `yaml:"build-idx"`
}

func ParseCbuildIndexFile(cbuildIndexFile string) (data CbuildIndex, err error) {
	yfile, err := os.ReadFile(cbuildIndexFile)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(yfile, &data)
	return
}
