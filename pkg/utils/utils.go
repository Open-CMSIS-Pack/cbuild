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

type ConfigurationItem struct {
	BuildType  string
	TargetType string
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
	parseError := errors.New("invalid context. Follow project.buildType+targetType symantic")

	periodCount := strings.Count(context, ".")
	plusCount := strings.Count(context, "+")
	if context == "" || periodCount > 1 || plusCount > 1 {
		err = parseError
		return
	}

	var projectName, buildType, targetType string

	targetIdx := strings.Index(context, "+")
	buildIdx := strings.Index(context, ".")

	if targetIdx == -1 || buildIdx == -1 {
		err = parseError
		return
	}

	// fully specified contexts
	part := context[:targetIdx]
	buildIdx = strings.Index(part, ".")

	if buildIdx > -1 {
		projectName = part[:buildIdx]
		buildType = part[buildIdx+1:]
	} else {
		projectName = part
	}

	part = context[targetIdx+1:]
	buildIdx = strings.Index(part, ".")

	if buildIdx > -1 {
		targetType = part[:buildIdx]
		buildType = part[buildIdx+1:]
	} else {
		targetType = part
	}

	if projectName == "" || buildType == "" || targetType == "" {
		err = parseError
		return
	}

	item.ProjectName = projectName
	item.BuildType = buildType
	item.TargetType = targetType
	return
}

func CreateConfiguration(configItem ConfigurationItem) (configuration string) {
	if configItem.BuildType != "" {
		configuration += "." + configItem.BuildType
	}
	if configItem.TargetType != "" {
		configuration += "+" + configItem.TargetType
	}
	return
}

func GetSelectedContexts(allContexts []string, configuration string) (selectedContexts []string, err error) {
	if configuration == "" {
		return selectedContexts, errors.New("configuration can not be empty")
	}

	configPattern := "^*" + configuration + "$"
	for _, context := range allContexts {
		match, regexErr := MatchToPattern(context, configPattern)
		if regexErr != nil {
			err = regexErr
			return
		}
		if match {
			selectedContexts = append(selectedContexts, context)
		}
	}

	if len(selectedContexts) == 0 {
		err = errors.New("specified configuration '" + configuration + "' not found")
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

func AppendUnique[T comparable](slice []T, elems ...T) []T {
	lookup := make(map[T]struct{})
	all := append(slice, elems...)
	var unique []T
	for _, elem := range all {
		_, isDuplicate := lookup[elem]
		if !isDuplicate {
			lookup[elem] = struct{}{}
			unique = append(unique, elem)
		}
	}
	return unique
}

func Contains[T comparable](slice []T, elem T) bool {
	for _, sliceElem := range slice {
		if sliceElem == elem {
			return true
		}
	}
	return false
}
