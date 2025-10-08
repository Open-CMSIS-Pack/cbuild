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
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// West Build Info
type WestBuildInfo struct {
	App        string
	Board      string
	OutDir     string
	Compiler   string
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

func GetModule(file string, modules []Module) string {
	name := "sources"
	path := ""
	for _, module := range modules {
		if strings.Contains(file, module.Path) {
			if len(module.Path) > len(path) {
				name = module.Name
				path = module.Path
			}
		}
	}
	return name
}

func AppendFileToGroup(fileTree *[]Filetree, group, file string) {
	for i := range *fileTree {
		if (*fileTree)[i].Group == group {
			(*fileTree)[i].Files = append((*fileTree)[i].Files, file)
			return
		}
	}
	*fileTree = append(*fileTree, Filetree{Group: group, Files: []string{file}})
}

func AddWestFilesToCbuild(westInfo WestBuildInfo) error {
	compileCommandsFile := filepath.Join(westInfo.OutDir, "compile_commands.json")
	compileCommandsData, _ := ParseCompileCommandsFile(compileCommandsFile)

	modulesFile := filepath.Join(westInfo.OutDir, "zephyr_modules.txt")
	modules, _ := ParseModules(modulesFile)

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
	var allFiles []string
	fileTree := []Filetree{} // make(map[string][]string)
	for _, compileCommands := range compileCommandsData {
		var file string
		file, err = filepath.Rel(filepath.Dir(westInfo.Cbuild), compileCommands.File)
		if err != nil {
			file = compileCommands.File
		}
		allFiles = append(allFiles, filepath.ToSlash(file))
		module := GetModule(filepath.ToSlash(compileCommands.File), modules)
		AppendFileToGroup(&fileTree, module, filepath.ToSlash(file))
	}

	// Find 'build' node
	buildNode := root.Content[0].Content[1]

	// Append groups
	groupsNode := &yaml.Node{Kind: yaml.SequenceNode}
	for _, module := range fileTree {
		groupNode := &yaml.Node{Kind: yaml.MappingNode}
		groupNode.Content = []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "group"},
			{Kind: yaml.ScalarNode, Value: module.Group},
		}
		filesNode := &yaml.Node{Kind: yaml.SequenceNode}
		for _, file := range module.Files {
			fileNode := &yaml.Node{Kind: yaml.MappingNode}
			fileNode.Content = []*yaml.Node{
				{Kind: yaml.ScalarNode, Value: "file"},
				{Kind: yaml.ScalarNode, Value: file},
			}
			filesNode.Content = append(filesNode.Content, fileNode)
		}
		groupNode.Content = append(groupNode.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "files"}, filesNode)
		groupsNode.Content = append(groupsNode.Content, groupNode)
	}
	buildNode.Content = append(buildNode.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "groups"}, groupsNode)

	// Append constructed-files as a workaround for populating the outline view
	constructedFilesNode := &yaml.Node{Kind: yaml.SequenceNode}
	for _, file := range allFiles {
		fileNode := &yaml.Node{Kind: yaml.MappingNode}
		fileNode.Content = []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "file"},
			{Kind: yaml.ScalarNode, Value: file},
		}
		constructedFilesNode.Content = append(constructedFilesNode.Content, fileNode)
	}
	buildNode.Content = append(buildNode.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "constructed-files"}, constructedFilesNode)

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
