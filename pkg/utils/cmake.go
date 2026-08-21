/*
 * Copyright (c) 2026 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
	"gopkg.in/yaml.v3"
)

type CMakeBuildInfo struct {
	OutDir string
	Cbuild string
}

func AddCMakeFilesToCbuild(cmakeInfo CMakeBuildInfo) error {
	compileCommandsFile := filepath.Join(cmakeInfo.OutDir, "compile_commands.json")
	if _, err := os.Stat(compileCommandsFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	fileTree, err := GetCompileCommandFileTree(cmakeInfo.OutDir, cmakeInfo.Cbuild, func(string) string {
		return CMakeSources
	})
	if err != nil {
		return err
	}

	data, err := os.ReadFile(cmakeInfo.Cbuild)
	if err != nil {
		return err
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return err
	}

	var buildNode *yaml.Node
	if len(root.Content) > 0 {
		buildNode = GetYamlNodeByKey(root.Content[0], "build")
	}
	if buildNode == nil {
		return errutils.New(errutils.ErrInvalidCbuildFormat, cmakeInfo.Cbuild)
	}

	groups := &yaml.Node{Kind: yaml.SequenceNode}
	for _, group := range fileTree {
		groupNode := AddGroup(groups, group.Group)
		AddFiles(groupNode, group.Files)
	}
	SetYamlNodeByKey(buildNode, groups, "groups")

	var buf strings.Builder
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(root.Content[0]); err != nil {
		return err
	}
	return os.WriteFile(cmakeInfo.Cbuild, []byte(buf.String()), 0600)
}
