/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testRoot = "../../test"

func TestGetExecutablePath(t *testing.T) {
	assert := assert.New(t)

	t.Run("get executable path", func(t *testing.T) {
		_, err := GetExecutablePath()
		assert.Nil(err)
	})
}

func TestUpdateEnvVars(t *testing.T) {
	assert := assert.New(t)

	t.Run("test update environment variables", func(t *testing.T) {
		binPath := testRoot + "/bin"
		etcPath := testRoot + "/etc"
		env := UpdateEnvVars(binPath, etcPath)
		binPath, _ = filepath.Abs(binPath)
		etcPath, _ = filepath.Abs(etcPath)
		assert.Equal(env.BuildRoot, binPath)
		assert.Equal(env.CompilerRoot, etcPath)
		assert.NotEmpty(env.PackRoot)
	})

	t.Run("test update environment variables, with CMSIS_PACK_ROOT", func(t *testing.T) {
		binPath := testRoot + "/bin"
		etcPath := testRoot + "/etc"
		packRoot, _ := filepath.Abs(testRoot + "/packs")
		_ = os.Setenv("CMSIS_PACK_ROOT", packRoot)
		env := UpdateEnvVars(binPath, etcPath)
		binPath, _ = filepath.Abs(binPath)
		etcPath, _ = filepath.Abs(etcPath)
		assert.Equal(env.BuildRoot, binPath)
		assert.Equal(env.CompilerRoot, etcPath)
		assert.NotEmpty(env.PackRoot)
	})
}

func TestGetDefaultCmsisPackRoot(t *testing.T) {
	assert := assert.New(t)

	t.Run("get default cmsis pack root", func(t *testing.T) {
		root := GetDefaultCmsisPackRoot()
		assert.NotEmpty(root)
	})
}

func TestParseContext(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		Input           string
		ExpectError     bool
		ExpectedContext ContextItem
	}{
		// negative test cases
		{"", true, ContextItem{}},
		{".+", true, ContextItem{}},
		{".Build+", true, ContextItem{}},
		{".+Target", true, ContextItem{}},
		{".Build.Build2+Target", true, ContextItem{}},
		{".Build+Target+Test", true, ContextItem{}},
		{"+Target", true, ContextItem{}},
		{".Build", true, ContextItem{}},
		{".Build+Target", true, ContextItem{}},
		{"+Target.Build", true, ContextItem{}},
		{"Project", true, ContextItem{}},
		{"Project.Build", true, ContextItem{}},
		{"Project.Build+", true, ContextItem{}},
		{"Project+Target.Build", true, ContextItem{}},

		// positive test cases
		{"Project.+Target", false, ContextItem{ProjectName: "Project", BuildType: "", TargetType: "Target"}},
		{"Project+Target", false, ContextItem{ProjectName: "Project", BuildType: "", TargetType: "Target"}},
		{"Project.Build+Target", false, ContextItem{ProjectName: "Project", BuildType: "Build", TargetType: "Target"}},
	}
	for _, test := range testCases {
		contextItem, err := ParseContext(test.Input)
		if test.ExpectError {
			assert.Error(err)
		} else {
			assert.Nil(err)
		}
		assert.Equal(contextItem.ProjectName, test.ExpectedContext.ProjectName)
		assert.Equal(contextItem.BuildType, test.ExpectedContext.BuildType)
		assert.Equal(contextItem.TargetType, test.ExpectedContext.TargetType)
	}
}

func TestCreateContext(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		Input          ContextItem
		ExpectError    bool
		ExpectedOutput string
	}{
		{ContextItem{ProjectName: "", BuildType: "", TargetType: ""}, true, ""},
		{ContextItem{ProjectName: "", BuildType: "Build", TargetType: "Target"}, true, ""},
		{ContextItem{ProjectName: "Project", BuildType: "Build", TargetType: ""}, true, ""},
		{ContextItem{ProjectName: "Project", BuildType: "", TargetType: "Target"}, false, "Project+Target"},
		{ContextItem{ProjectName: "Project", BuildType: "Build", TargetType: "Target"}, false, "Project.Build+Target"},
	}
	for _, test := range testCases {
		context, err := CreateContext(test.Input)
		if test.ExpectError {
			assert.Error(err)
		} else {
			assert.Nil(err)
		}
		assert.Equal(context, test.ExpectedOutput)
	}
}

func TestParseConfiguration(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		Input           string
		ExpectError     bool
		ExpectedContext ConfigurationItem
	}{
		{"", true, ConfigurationItem{}},
		{".+", true, ConfigurationItem{}},
		{"Project", true, ConfigurationItem{}},
		{".Build.Build2+Target", true, ConfigurationItem{}},
		{".Build+Target+Test", true, ConfigurationItem{}},
		{"Project.Build", true, ConfigurationItem{}},
		{"Project+Target", true, ConfigurationItem{}},
		{"Project.Build+Target", true, ConfigurationItem{}},
		{"Project+Target.Build", true, ConfigurationItem{}},
		{".+Target", false, ConfigurationItem{BuildType: "", TargetType: "Target"}},
		{".Build+", false, ConfigurationItem{BuildType: "Build", TargetType: ""}},
		{"+Target", false, ConfigurationItem{BuildType: "", TargetType: "Target"}},
		{".Build", false, ConfigurationItem{BuildType: "Build", TargetType: ""}},
		{".Build+Target", false, ConfigurationItem{BuildType: "Build", TargetType: "Target"}},
		{"+Target.Build", false, ConfigurationItem{BuildType: "Build", TargetType: "Target"}},
	}
	for _, test := range testCases {
		configItem, err := ParseConfiguration(test.Input)
		if test.ExpectError {
			assert.Error(err)
		} else {
			assert.Nil(err)
		}
		assert.Equal(configItem.BuildType, test.ExpectedContext.BuildType)
		assert.Equal(configItem.TargetType, test.ExpectedContext.TargetType)
	}
}

func TestCreateConfiguration(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		Input          ConfigurationItem
		ExpectedOutput string
	}{
		{ConfigurationItem{}, ""},
		{ConfigurationItem{BuildType: "Build", TargetType: "Target"}, ".Build+Target"},
		{ConfigurationItem{BuildType: "", TargetType: "Target"}, "+Target"},
		{ConfigurationItem{BuildType: "Build", TargetType: ""}, ".Build"},
	}
	for _, test := range testCases {
		config := CreateConfiguration(test.Input)
		assert.Equal(config, test.ExpectedOutput)
	}
}

func TestGetSelectedContexts(t *testing.T) {
	assert := assert.New(t)
	var empty []string

	allContexts := []string{
		"Project.Debug+Target1",
		"Project.Debug+Target2",
		"Project.Release+Target1",
		"Project.Release+Target2",
	}

	testCases := []struct {
		InputContext             string
		ExpectError              bool
		ExpectedSelectedContexts []string
	}{
		{"", true, empty},
		{"UnknowProject+TestTarget.TestBuild", true, empty},
		{"Project+Target1.Build.Release", true, empty},
		{"Project+Target1+Target2.Release", true, empty},
		{"Project", true, empty},
		{"Project.", true, empty},
		{"Project+", true, empty},
		{"Project.+", true, empty},
		{"Project+Target1", true, empty},
		{"Project.Release", true, empty},
		{"Project.Debug+Target1", true, empty},
		{"Project+Target1.Debug", true, empty},
		{"+Target1", false, []string{"Project.Debug+Target1", "Project.Release+Target1"}},
		{".Debug", false, []string{"Project.Debug+Target1", "Project.Debug+Target2"}},
		{".Debug+Target1", false, []string{"Project.Debug+Target1"}},
		{"+Target1.Debug", false, []string{"Project.Debug+Target1"}},
	}
	for _, test := range testCases {
		selectedContexts, err := GetSelectedContexts(allContexts, test.InputContext)
		if test.ExpectError {
			assert.Error(err)
		} else {
			assert.Nil(err)
		}
		assert.Equal(selectedContexts, test.ExpectedSelectedContexts)
	}
}

func TestParseCbuildIndexFile(t *testing.T) {
	assert := assert.New(t)

	t.Run("test file not available", func(t *testing.T) {
		_, err := ParseCbuildIndexFile("Unknown.cbuild-idx.yml")
		assert.Error(err)
	})

	t.Run("test", func(t *testing.T) {
		data, err := ParseCbuildIndexFile(testRoot + "/run/Test.cbuild-idx.yml")
		assert.Nil(err)
		assert.Equal(data.BuildIdx.GeneratedBy, "csolution 1.4.0")
		assert.Equal(data.BuildIdx.Cdefault, "HelloWorld.cdefault.yml")
		assert.Equal(data.BuildIdx.Csolution, "HelloWorld.csolution.yml")
		assert.Equal(len(data.BuildIdx.Cprojects), 2)
		assert.Equal(data.BuildIdx.Cprojects[0].Cproject, "cm0plus/HelloWorld_cm0plus.cproject.yml")
		assert.Equal(data.BuildIdx.Cprojects[1].Cproject, "cm4/HelloWorld_cm4.cproject.yml")
		assert.Equal(data.BuildIdx.Licenses, "test123")
		assert.Equal(len(data.BuildIdx.Cbuilds), 4)
		assert.Equal(data.BuildIdx.Cbuilds[0].Cbuild, "cm0plus/HelloWorld_cm0plus.Debug+FRDM-K32L3A6.cbuild.yml")
		assert.Equal(data.BuildIdx.Cbuilds[1].Cbuild, "cm0plus/HelloWorld_cm0plus.Release+FRDM-K32L3A6.cbuild.yml")
		assert.Equal(data.BuildIdx.Cbuilds[2].Cbuild, "cm4/HelloWorld_cm4.Debug+FRDM-K32L3A6.cbuild.yml")
		assert.Equal(data.BuildIdx.Cbuilds[3].Cbuild, "cm4/HelloWorld_cm4.Release+FRDM-K32L3A6.cbuild.yml")
	})
}

func TestAppendUnique(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		input            []string
		addElement       string
		expectedSliceLen int
		expectedOutput   []string
	}{
		{[]string{"one", "two", "three"}, "four", 4, []string{"one", "two", "three", "four"}},
		{[]string{"one", "two", "three"}, "one", 3, []string{"one", "two", "three"}},
	}
	for _, test := range testCases {
		output := AppendUnique(test.input, test.addElement)
		assert.Len(output, test.expectedSliceLen)
		assert.Equal(output, test.expectedOutput)
	}
}

func TestContains(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		slice          []string
		element        string
		expectedResult bool
	}{
		{[]string{"one", "two", "three"}, "four", false},
		{[]string{""}, "one", false},
		{[]string{"one", "two", "three"}, "one", true},
	}

	for _, test := range testCases {
		output := Contains(test.slice, test.element)
		assert.Equal(output, test.expectedResult)
	}
}
