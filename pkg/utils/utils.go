/*
 * Copyright (c) 2022-2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
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
	parseError := errutils.New(errutils.ErrInvalidContextFormat)

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
		TmpDir      string `yaml:"tmpdir"`
		Cprojects   []struct {
			Cproject string `yaml:"cproject"`
		} `yaml:"cprojects"`
		Licenses interface{} `yaml:"licenses"`
		Cbuilds  []struct {
			Cbuild        string `yaml:"cbuild"`
			Project       string `yaml:"project"`
			Configuration string `yaml:"configuration"`
			Rebuild       bool   `yaml:"rebuild"`
		} `yaml:"cbuilds"`
		Executes []interface{} `yaml:"executes"`
		Rebuild  bool          `yaml:"rebuild"`
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

func ResolveContexts(allContext []string, contextFilters []string) ([]string, error) {
	var selectedContexts []string

	// remove duplicates (if any)
	filters := RemoveDuplicates(contextFilters)

	for _, filter := range filters {
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
			return nil, errutils.New(errutils.ErrNoFilteredContextFound, filter)
		}
	}
	return selectedContexts, nil
}

func LogStdMsg(msg string) {
	if msg != "" {
		_, _ = log.StandardLogger().Out.Write([]byte(msg + "\n"))
	}
}

func FormatTime(time time.Duration) string {
	// Format time in "hh:mm:ss"
	return fmt.Sprintf("%02d:%02d:%02d", int(time.Hours()), int(time.Minutes())%60, int(time.Seconds())%60)
}

func RemoveDuplicates(input []string) []string {
	// Create a map to track seen strings
	seen := make(map[string]bool)
	// Create a slice to store the unique strings
	var result []string

	// Iterate over the input slice
	for _, str := range input {
		// If the string is not in the map,
		// add it to the result and mark it as seen
		if !seen[str] {
			result = append(result, str)
			seen[str] = true
		}
	}

	return result
}

func FileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		// File exists
		return true, nil
	}
	if os.IsNotExist(err) {
		// File doesn't exist
		return false, errutils.New(errutils.ErrFileNotExist, filePath)
	}
	// Return error for any other issues (permission denied, etc.)
	return false, err
}

func PrintSeparator(delimiter string, length int) {
	if length > 0 {
		sep := strings.Repeat(delimiter, length-1)
		LogStdMsg("+" + sep)
	}
}

// checks if two paths are equivalent
func ComparePaths(path1, path2 string) (bool, error) {
	cleanPath1 := filepath.Clean(path1)
	cleanPath2 := filepath.Clean(path2)

	absPath1, err := filepath.Abs(cleanPath1)
	if err != nil {
		return false, err
	}
	absPath2, err := filepath.Abs(cleanPath2)
	if err != nil {
		return false, err
	}

	if isFileSystemCaseInsensitive() {
		absPath1 = strings.ToLower(absPath1)
		absPath2 = strings.ToLower(absPath2)
	}

	return absPath1 == absPath2, nil
}

func isFileSystemCaseInsensitive() bool {
	// On Windows and macOS, file systems are typically case insensitive
	// On Linux, file systems are typically case sensitive
	return filepath.Separator == '\\' || strings.Contains(strings.ToLower(os.Getenv("OS")), "darwin")
}
