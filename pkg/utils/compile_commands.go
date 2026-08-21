/*
 * Copyright (c) 2025-2026 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v3"
)

const CMakeSources = "CMake Sources"

type CompileCommands struct {
	Directory string `json:"directory"`
	File      string `json:"file"`
	Output    string `json:"output"`
	Command   string `json:"command"`
}

type Filetree struct {
	Group string
	Files []string
}

func ParseCompileCommandsFile(compileCommandsFile string) ([]CompileCommands, error) {
	var data []CompileCommands
	err := ParseYAMLFile(compileCommandsFile, &data)
	return data, err
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

func GetCompileCommandFileTree(outDir, cbuild string, getGroup func(string) string) ([]Filetree, error) {
	compileCommandsFile := filepath.Join(outDir, "compile_commands.json")
	compileCommandsData, err := ParseCompileCommandsFile(compileCommandsFile)
	if err != nil {
		return nil, err
	}

	fileTree := []Filetree{}
	for _, compileCommand := range compileCommandsData {
		file, err := filepath.Rel(filepath.Dir(cbuild), compileCommand.File)
		if err != nil {
			file = compileCommand.File
		}
		AppendFileToGroupUniquely(&fileTree, getGroup(filepath.ToSlash(compileCommand.File)), filepath.ToSlash(file))
	}
	return fileTree, nil
}
