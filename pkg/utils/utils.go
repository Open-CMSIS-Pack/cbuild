/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"errors"
	"os"
	"os/exec"
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
	env.BuildRoot, _ = filepath.Abs(binPath)
	log.Debug("CMSIS_PACK_ROOT: " + env.PackRoot)
	log.Debug("CMSIS_COMPILER_ROOT: " + env.CompilerRoot)
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
	parseError := errors.New("invalid context. Follow [project][.buildType][+targetType] syntax")

	periodCount := strings.Count(context, ".")
	plusCount := strings.Count(context, "+")
	if context == "" || periodCount > 1 || plusCount > 1 {
		err = parseError
		return
	}

	var projectName, buildType, targetType string

	targetIdx := strings.Index(context, "+")
	buildIdx := strings.Index(context, ".")

	if (targetIdx != -1 && buildIdx != -1) && targetIdx < buildIdx {
		err = parseError
		return
	}

	if targetIdx == -1 && buildIdx == -1 {
		projectName = context
	} else if buildIdx == -1 {
		// context with only projectName+targetType
		projectName = context[:targetIdx]
		targetType = context[targetIdx+1:]
	} else if targetIdx == -1 {
		// context with only projectName.buildtype
		projectName = context[:buildIdx]
		buildType = context[buildIdx+1:]
	} else {
		// fully specified contexts
		part := context[:targetIdx]
		buildIdx := strings.Index(part, ".")

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
	}

	item.ProjectName = projectName
	item.BuildType = buildType
	item.TargetType = targetType
	return
}

func CreateContext(contextItem ContextItem) (context string) {
	context = contextItem.ProjectName
	if contextItem.BuildType != "" {
		context += "." + contextItem.BuildType
	}
	if contextItem.TargetType != "" {
		context += "+" + contextItem.TargetType
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
			Cbuild        string `yaml:"cbuild"`
			Project       string `json:"project"`
			Configuration string `json:"configuration"`
		} `yaml:"cbuilds"`
	} `yaml:"build-idx"`
}

type CbuildSet struct {
	ContextSet struct {
		GeneratedBy string `yaml:"generated-by"`
		Contexts    []struct {
			Context string `yaml:"context"`
		} `yaml:"contexts"`
		Compiler string `yaml:"compiler"`
	} `yaml:"cbuild-set"`
}

func ParseCbuildIndexFile(cbuildIndexFile string) (data CbuildIndex, err error) {
	yfile, err := os.ReadFile(cbuildIndexFile)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(yfile, &data)
	return
}

func ParseCbuildSetFile(cbuildSetFile string) (data CbuildSet, err error) {
	yfile, err := os.ReadFile(cbuildSetFile)
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

func GetInstalledExePath(exeName string) (path string, err error) {
	path, err = exec.LookPath(exeName)
	if err != nil {
		path = NormalizePath(path)
	}
	return
}

func NormalizePath(path string) string {
	if strings.Contains(path, "\\") {
		path = strings.ReplaceAll(path, "\\", "/")
	}
	return path
}

func GetProjectName(csolutionFile string) (projectName string, err error) {
	csolutionFile = NormalizePath(csolutionFile)
	nameTokens := strings.Split(filepath.Base(csolutionFile), ".")
	if len(nameTokens) != 3 {
		return "", errors.New("invalid csolution file name")
	}
	return nameTokens[0], nil
}

func ResolveContexts(allContext []string, contextFilters []string) ([]string, error) {
	var selectedContexts []string

	for _, filter := range contextFilters {
		filterContextItem, err := ParseContext(filter)
		if err != nil {
			return nil, err
		}
		matchFound := false
		for _, context := range allContext {
			availableContextItem, err := ParseContext(context)
			if err != nil {
				return nil, err
			}

			var contextPattern string
			if filterContextItem.ProjectName != "" {
				contextPattern = filterContextItem.ProjectName
			} else {
				contextPattern = "*"
			}

			contextPattern += "."
			if filterContextItem.BuildType != "" {
				contextPattern += filterContextItem.BuildType
			} else {
				contextPattern += "*"
			}

			contextPattern += "+"
			if filterContextItem.TargetType != "" {
				contextPattern += filterContextItem.TargetType
			} else {
				contextPattern += "*"
			}

			fullContextItem := availableContextItem.ProjectName + "." + availableContextItem.BuildType + "+" + availableContextItem.TargetType

			match, err := MatchString(fullContextItem, contextPattern)
			if err != nil {
				return nil, err
			}
			if match && !Contains(selectedContexts, context) {
				matchFound = match
				selectedContexts = append(selectedContexts, context)
			}
		}
		if !matchFound {
			return nil, errors.New("no suitable match found for \"" + filter + "\"")
		}
	}
	return selectedContexts, nil
}
