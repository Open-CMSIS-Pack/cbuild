/*
 * Copyright (c) 2022-2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
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
			West          bool   `yaml:"west"`
			Project       string `yaml:"project"`
			Configuration string `yaml:"configuration"`
			Rebuild       bool   `yaml:"rebuild"`
			Messages      struct {
				Warnings []string `yaml:"warnings"`
				Info     []string `yaml:"info"`
			} `yaml:"messages"`
		} `yaml:"cbuilds"`
		Executes  []interface{} `yaml:"executes"`
		Rebuild   bool          `yaml:"rebuild"`
		ImageOnly bool          `yaml:"image-only"`
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

type Csolution struct {
	Solution struct {
		TargetTypes []struct {
			Type      string `yaml:"type"`
			TargetSet []struct {
				Set    string `yaml:"set"`
				Images []struct {
					ProjectContext string `yaml:"project-context"`
				} `yaml:"images"`
			} `yaml:"target-set"`
		} `yaml:"target-types"`
		OutputDirs struct {
			Tmpdir string `yaml:"tmpdir"`
		} `yaml:"output-dirs"`
	} `yaml:"solution"`
}

type Cbuild struct {
	Build struct {
		OutputDirs struct {
			Intdir string `yaml:"intdir"`
			Outdir string `yaml:"outdir"`
		} `yaml:"output-dirs"`
	} `yaml:"build"`
}

func ParseYAMLFile(filePath string, out interface{}) error {
	// Read the file
	yfile, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Unmarshal the YAML into the provided structure
	err = yaml.Unmarshal(yfile, out)
	return err
}

// Wrapper functions for specific types
func ParseCbuildIndexFile(cbuildIndexFile string) (CbuildIndex, error) {
	var data CbuildIndex
	err := ParseYAMLFile(cbuildIndexFile, &data)
	return data, err
}

func ParseCbuildSetFile(cbuildSetFile string) (CbuildSet, error) {
	var data CbuildSet
	err := ParseYAMLFile(cbuildSetFile, &data)
	return data, err
}

func ParseCsolutionFile(csolutionFile string) (Csolution, error) {
	var data Csolution
	err := ParseYAMLFile(csolutionFile, &data)
	return data, err
}

func ParseCbuildFile(cbuildFile string) (Cbuild, error) {
	var data Cbuild
	err := ParseYAMLFile(cbuildFile, &data)
	return data, err
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

func GetTmpDir(csolutionFile string, outputDir string) (string, error) {
	// Default temporary directory name
	const defaultTmpDir = "tmp"

	// Get the base directory of the csolution file
	basePath := filepath.Dir(csolutionFile)

	// Parse the csolution file
	data, err := ParseCsolutionFile(csolutionFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Return an error if the csolution file does not exist
			return "", err
		}

		// For other parsing errors, fallback to the default tmp directory
		tmpPath := filepath.Join(basePath, outputDir, defaultTmpDir)
		return NormalizePath(tmpPath), nil
	}

	tmpDir := data.Solution.OutputDirs.Tmpdir
	if tmpDir == "" {
		tmpDir = defaultTmpDir
	}

	tmpPath := filepath.Join(basePath, outputDir, tmpDir)
	return NormalizePath(tmpPath), nil
}

func GetOutDir(cbuildIdxFile string, context string) (string, error) {
	basePath := filepath.Dir(cbuildIdxFile)
	defaultOutPath := filepath.Join(basePath, "out")

	// Check if the cbuild index file exists
	if _, err := os.Stat(cbuildIdxFile); os.IsNotExist(err) {
		return defaultOutPath, nil
	}

	// Parse the cbuild index file
	data, err := ParseCbuildIndexFile(cbuildIdxFile)
	if err != nil {
		return "", err
	}

	// Locate the cbuild file based on the provided context
	var cbuildFile string
	for _, cbuild := range data.BuildIdx.Cbuilds {
		if context == cbuild.Project+cbuild.Configuration {
			cbuildFile = cbuild.Cbuild
			break
		}
	}

	if cbuildFile == "" {
		return defaultOutPath, nil // Fallback to default if no match found
	}

	cbuildFilePath := filepath.Join(basePath, cbuildFile)

	// Parse the cbuild file
	cbuildData, err := ParseCbuildFile(cbuildFilePath)
	if err != nil {
		return "", err
	}

	// Determine the output directory
	outDir := cbuildData.Build.OutputDirs.Outdir
	if outDir == "" {
		return defaultOutPath, nil
	}

	// Resolve relative paths to absolute
	if !filepath.IsAbs(outDir) {
		return filepath.Join(filepath.Dir(cbuildFilePath), outDir), nil
	}

	return outDir, nil
}

// matchAnyPattern checks if a filename matches any of the provided glob patterns
func matchAnyPattern(filename string, patterns []string) (bool, error) {
	for _, pattern := range patterns {
		match, err := filepath.Match(pattern, filename)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

// DeleteAll removes everything under path except files whose base name
// matches any of the provided glob patterns. If patterns is empty, it just calls os.RemoveAll.
func DeleteAll(path string, excludeFilePatterns []string) error {
	// check path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errutils.New(errutils.ErrPathNotExist, path)
	}

	// if no patterns given, just delete everything
	if len(excludeFilePatterns) == 0 {
		if err := os.RemoveAll(path); err != nil {
			return errutils.New(errutils.ErrDeleteFailed, path)
		}
		return nil
	}

	// Walk the tree and delete all non-matching files
	var walkErr error
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			// problem accessing p
			return err
		}
		// skip root itself
		if p == path {
			return nil
		}
		// if this is a file…
		if !info.IsDir() {
			// check if file matches any of the exclude patterns
			shouldExclude, merr := matchAnyPattern(info.Name(), excludeFilePatterns)
			if merr != nil {
				return merr
			}

			if shouldExclude {
				// do not delete matching file
				return nil
			}

			// delete non-matching file
			if derr := os.Remove(p); derr != nil {
				// collect error but keep going
				walkErr = derr
			}
		}
		return nil
	})
	if err != nil {
		return errutils.New(errutils.ErrDeleteFailed, err.Error())
	}
	if walkErr != nil {
		return errutils.New(errutils.ErrDeleteFailed, walkErr.Error())
	}

	// Collect all directories
	var dirs []string
	_ = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			dirs = append(dirs, p)
		}
		return nil
	})

	// Sort so deepest directories come first
	sort.Slice(dirs, func(path1, path2 int) bool {
		return len(dirs[path1]) > len(dirs[path2])
	})

	// Try removing each if it’s now empty
	for _, dir := range dirs {
		// skip the root; we won't remove path itself
		if dir == path {
			continue
		}
		// if directory is now empty, delete it
		if entries, rerr := os.ReadDir(dir); rerr == nil && len(entries) == 0 {
			if dirErr := os.Remove(dir); dirErr != nil {
				// collect but don't abort
				walkErr = dirErr
			}
		}
	}
	if walkErr != nil {
		return errutils.New(errutils.ErrDeleteFailed, walkErr.Error())
	}

	return nil
}

func ParseAndFetchToolchainInfo(toolchainFile string) string {
	// Open the toolchain.cmake file
	file, err := os.Open(toolchainFile)
	if err != nil {
		return ""
	}
	defer file.Close()

	// regex patterns to extract the required information
	rootPattern := `set\(REGISTERED_TOOLCHAIN_ROOT\s+"([^"]+)"\)`
	versionPattern := `set\(REGISTERED_TOOLCHAIN_VERSION\s+"([^"]+)"\)`
	compilerPattern := `include\("\${CMSIS_COMPILER_ROOT}/(.*)\.\d+\.\d+\.\d+\.cmake"\)`

	var toolchainRoot, toolchainVersion, compilerName string
	// Scan toolchain.cmake file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Get toolchain root
		if matches := regexp.MustCompile(rootPattern).FindStringSubmatch(line); len(matches) > 1 {
			toolchainRoot = matches[1]
		}

		// Get matched toolchain version
		if matches := regexp.MustCompile(versionPattern).FindStringSubmatch(line); len(matches) > 1 {
			toolchainVersion = matches[1]
		}

		// Get matched toolchain name
		if matches := regexp.MustCompile(compilerPattern).FindStringSubmatch(line); len(matches) > 1 {
			compilerName = strings.Split(matches[1], ".")[0]
		}
	}

	// Check all required values were found
	if toolchainRoot == "" || toolchainVersion == "" || compilerName == "" {
		return ""
	}

	return fmt.Sprintf("Using %s V%s compiler, from: '%s'", compilerName, toolchainVersion, toolchainRoot)
}

func GetParentFolder(path string) (string, error) {
	if path == "" {
		return "", errutils.New(errutils.ErrInvalidPath, "empty path provided")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", errutils.New(errutils.ErrFetchingAbsPath, err.Error())
	}

	parentPath := filepath.Dir(absPath)
	return filepath.Base(parentPath), nil
}

func GetTargetSetProjectContexts(csolutionFile string, selectedTargetSet string) []string {
	targetType, targetSet, _ := strings.Cut(selectedTargetSet, "@")

	// Parse the csolution file
	data, _ := ParseCsolutionFile(csolutionFile)

	// Get project contexts for the selected target set
	for _, tt := range data.Solution.TargetTypes {
		if tt.Type == targetType {
			for _, ts := range tt.TargetSet {
				if ts.Set == targetSet {
					var contexts []string
					for _, img := range ts.Images {
						contexts = append(contexts, img.ProjectContext+"+"+targetType)
					}
					return contexts
				}
			}
		}
	}
	return []string{}
}
