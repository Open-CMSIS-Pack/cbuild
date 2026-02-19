/*
 * Copyright (c) 2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"bytes"
	"encoding/csv"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
	log "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/logger"
	"gopkg.in/yaml.v3"
)

// Group names
const App = "App"
const Generated = "Generated"
const ZephyrModules = "Zephyr Modules"
const ZephyrSources = "Zephyr Sources"

// West Build Info
type WestBuildInfo struct {
	AppPath    string
	OutDir     string
	Cbuild     string
	CbuildData Cbuild
}

// Compile Commands
type CompileCommands struct {
	Directory string `json:"directory"`
	File      string `json:"file"`
	Output    string `json:"output"`
	Command   string `json:"command"`
}

// West Modules
type Module struct {
	Name  string
	Path  string
	CMake string
}

// West Groups
type Filetree struct {
	Group string
	Files []string
}

func ParseCompileCommandsFile(compileCommandsFile string) ([]CompileCommands, error) {
	var data []CompileCommands
	err := ParseYAMLFile(compileCommandsFile, &data)
	return data, err
}

func ParseModules(filePath string) ([]Module, error) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	cr := csv.NewReader(bytes.NewReader(fileBytes))
	cr.Comma = ':'
	cr.FieldsPerRecord = 3
	cr.LazyQuotes = true
	cr.TrimLeadingSpace = true
	cr.Comment = '#'

	var out []Module
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		out = append(out, Module{
			Name:  rec[0],
			Path:  rec[1],
			CMake: rec[2],
		})
	}
	return out, nil
}

func GetGroupNames(module string) []string {
	// Group names according to module name
	if len(module) == 0 {
		return []string{ZephyrSources}
	} else if module == App || module == Generated {
		return []string{module}
	} else {
		return []string{ZephyrModules, module}
	}
}

func GetModule(file string, modules []Module) string {
	var name, path string
	file = strings.ToLower(file)
	for _, module := range modules {
		if strings.Contains(file, strings.ToLower(module.Path)) {
			if len(module.Path) > len(path) {
				name = module.Name
				path = module.Path
			}
		}
	}
	return name
}

func AppendFileToGroupUniquely(fileTree *[]Filetree, group, file string) {
	for i := range *fileTree {
		if (*fileTree)[i].Group == group {
			if !slices.Contains((*fileTree)[i].Files, file) {
				(*fileTree)[i].Files = append((*fileTree)[i].Files, file)
			}
			return
		}
	}
	*fileTree = append(*fileTree, Filetree{Group: group, Files: []string{file}})
}

func GetYamlNodeByKey(node *yaml.Node, key string) *yaml.Node {
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func SetYamlNodeByKey(base *yaml.Node, node *yaml.Node, key string) {
	p := GetYamlNodeByKey(base, key)
	if p != nil {
		*p = *node
	} else {
		base.Content = append(base.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: key}, node)
	}
}

func SetYamlNodeKeyValue(node *yaml.Node, key string, value string) {
	node.Content = []*yaml.Node{
		{Kind: yaml.ScalarNode, Value: key},
		{Kind: yaml.ScalarNode, Value: value},
	}
}

func AddFiles(parent *yaml.Node, files []string) {
	filesNode := &yaml.Node{Kind: yaml.SequenceNode}
	for _, file := range files {
		fileNode := &yaml.Node{Kind: yaml.MappingNode}
		SetYamlNodeKeyValue(fileNode, "file", file)
		filesNode.Content = append(filesNode.Content, fileNode)
	}
	parent.Content = append(parent.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "files"}, filesNode)
}

func AddGroup(parent *yaml.Node, group string) *yaml.Node {
	groupNode := &yaml.Node{Kind: yaml.MappingNode}
	SetYamlNodeKeyValue(groupNode, "group", group)
	parent.Content = append(parent.Content, groupNode)
	return groupNode
}

func AddGroupsAndFiles(node *yaml.Node, zephyr *yaml.Node, groups []string, files []string) {
	var group *yaml.Node
	if groups[0] == ZephyrModules {
		group = AddGroup(zephyr, groups[1])
	} else {
		group = AddGroup(node, groups[0])
	}
	AddFiles(group, files)
}

func AddWestFilesToCbuild(westInfo WestBuildInfo) error {
	compileCommandsFile := filepath.Join(westInfo.OutDir, "compile_commands.json")
	compileCommandsData, _ := ParseCompileCommandsFile(compileCommandsFile)

	modulesFile := filepath.Join(westInfo.OutDir, "zephyr_modules.txt")
	modules, _ := ParseModules(modulesFile)
	modules = append(modules, Module{Name: App, Path: westInfo.AppPath})
	modules = append(modules, Module{Name: Generated, Path: westInfo.OutDir})

	// Read Cbuild file
	data, err := os.ReadFile(westInfo.Cbuild)
	if err != nil {
		panic(err)
	}

	// Parse into a generic map to preserve order
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		panic(err)
	}

	// Get all files and separate them by modules
	fileTree := []Filetree{}
	for _, compileCommands := range compileCommandsData {
		var file string
		file, err = filepath.Rel(filepath.Dir(westInfo.Cbuild), compileCommands.File)
		if err != nil {
			file = compileCommands.File
		}
		module := GetModule(filepath.ToSlash(compileCommands.File), modules)
		AppendFileToGroupUniquely(&fileTree, module, filepath.ToSlash(file))
	}

	// Find 'build' node
	var buildNode *yaml.Node
	if len(root.Content) > 0 {
		buildNode = GetYamlNodeByKey(root.Content[0], "build")
	}
	if buildNode == nil {
		err := errutils.New(errutils.ErrInvalidCbuildFormat, westInfo.Cbuild)
		return err
	}

	// Create groups
	groups := &yaml.Node{Kind: yaml.SequenceNode}
	zephyr := &yaml.Node{Kind: yaml.SequenceNode}
	for _, module := range fileTree {
		AddGroupsAndFiles(groups, zephyr, GetGroupNames(module.Group), module.Files)
	}
	zephyrGroup := AddGroup(groups, ZephyrModules)
	zephyrGroup.Content = append(zephyrGroup.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "groups"}, zephyr)

	// Replace 'groups' node if it exists, otherwise append it
	SetYamlNodeByKey(buildNode, groups, "groups")

	// Update Cbuild file
	var buf strings.Builder
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(root.Content[0]); err != nil {
		panic(err)
	}
	file, err := os.Create(westInfo.Cbuild)
	if err == nil {
		_, _ = file.WriteString(buf.String())
	}
	file.Close()

	return nil
}

func CheckEnvVars(vars []string) {
	// Look for environment variables
	for _, key := range vars {
		value := os.Getenv(key)
		if value == "" {
			log.Warn("missing " + key + " environment variable")
			// #nosec G703 value comes from environment variable (deployment config)
		} else if _, err := os.Stat(value); os.IsNotExist(err) {
			log.Warn(key + " environment variable specifies non-existent directory: " + value)
		}
		log.Debug(key + "=" + value)
	}
}

func CheckWestSetup() error {
	// Check environment variables
	CheckEnvVars([]string{"ZEPHYR_BASE", "VIRTUAL_ENV"})
	// Look for west binary
	westBin, err := exec.LookPath("west")
	if err != nil {
		return err
	}
	log.Debug("west found: " + westBin)
	return nil
}
