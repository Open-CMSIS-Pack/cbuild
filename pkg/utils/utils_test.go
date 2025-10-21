/*
 * Copyright (c) 2022-2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/errutils"
	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/inittest"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../test"
const testDir = "utils"

func init() {
	inittest.TestInitialization(testRoot, testDir)
}

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
		//{".+", true, ContextItem{}},

		{".Build.Build2+Target", true, ContextItem{}},
		{".Build+Target+Test", true, ContextItem{}},
		{"+Target.Build", true, ContextItem{}},
		{"Project+Target.Build", true, ContextItem{}},

		// positive test cases
		{".Build+", false, ContextItem{ProjectName: "", BuildType: "Build", TargetType: ""}},
		{".+Target", false, ContextItem{ProjectName: "", BuildType: "", TargetType: "Target"}},
		{"+Target", false, ContextItem{ProjectName: "", BuildType: "", TargetType: "Target"}},
		{".Build", false, ContextItem{ProjectName: "", BuildType: "Build", TargetType: ""}},
		{".Build+Target", false, ContextItem{ProjectName: "", BuildType: "Build", TargetType: "Target"}},
		{"Project", false, ContextItem{ProjectName: "Project", BuildType: "", TargetType: ""}},
		{"Project.Build", false, ContextItem{ProjectName: "Project", BuildType: "Build", TargetType: ""}},
		{"Project.Build+", false, ContextItem{ProjectName: "Project", BuildType: "Build", TargetType: ""}},
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
		{ContextItem{ProjectName: "", BuildType: "", TargetType: ""}, false, ""},
		{ContextItem{ProjectName: "Project", BuildType: "", TargetType: ""}, false, "Project"},
		{ContextItem{ProjectName: "", BuildType: "Build", TargetType: ""}, false, ".Build"},
		{ContextItem{ProjectName: "", BuildType: "", TargetType: "Target"}, false, "+Target"},
		{ContextItem{ProjectName: "Project", BuildType: "Build", TargetType: ""}, false, "Project.Build"},
		{ContextItem{ProjectName: "", BuildType: "Build", TargetType: "Target"}, false, ".Build+Target"},
		{ContextItem{ProjectName: "Project", BuildType: "Build", TargetType: ""}, false, "Project.Build"},
		{ContextItem{ProjectName: "Project", BuildType: "", TargetType: "Target"}, false, "Project+Target"},
		{ContextItem{ProjectName: "Project", BuildType: "Build", TargetType: "Target"}, false, "Project.Build+Target"},
	}
	for _, test := range testCases {
		context := CreateContext(test.Input)
		assert.Equal(context, test.ExpectedOutput)
	}
}

func TestParseCbuildIndexFile(t *testing.T) {
	assert := assert.New(t)

	t.Run("test file not available", func(t *testing.T) {
		_, err := ParseCbuildIndexFile("Unknown.cbuild-idx.yml")
		assert.Error(err)
	})

	t.Run("test cbuild-idx file parsing", func(t *testing.T) {
		data, err := ParseCbuildIndexFile(filepath.Join(testRoot, testDir, "Test.cbuild-idx.yml"))
		assert.Nil(err)
		var re = regexp.MustCompile(`^csolution\s[\d]+.[\d+]+.[\d+].*`)
		assert.True(re.MatchString(data.BuildIdx.GeneratedBy))
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

func TestParseCbuildSetFile(t *testing.T) {
	assert := assert.New(t)

	t.Run("test file not available", func(t *testing.T) {
		_, err := ParseCbuildSetFile("Unknown.cbuild-idx.yml")
		assert.Error(err)
	})

	t.Run("test cbuild-set file parsing", func(t *testing.T) {
		data, err := ParseCbuildSetFile(filepath.Join(testRoot, testDir, "Test.cbuild-set.yml"))
		assert.Nil(err)
		var re = regexp.MustCompile(`^csolution\sversion\s[\d]+.[\d+]+.[\d+].*`)
		assert.True(re.MatchString(data.ContextSet.GeneratedBy))
		assert.Equal(len(data.ContextSet.Contexts), 4)
		assert.Equal(data.ContextSet.Contexts[0].Context, "test2.Debug+CM0")
		assert.Equal(data.ContextSet.Contexts[1].Context, "test2.Debug+CM3")
		assert.Equal(data.ContextSet.Contexts[2].Context, "test1.Debug+CM0")
		assert.Equal(data.ContextSet.Contexts[3].Context, "test1.Release+CM0")
		assert.Equal(data.ContextSet.Compiler, "AC6")
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

func TestGetInstalledExePath(t *testing.T) {
	assert := assert.New(t)
	t.Run("test to get invalid executable path", func(t *testing.T) {
		path, err := GetInstalledExePath("testunknown")
		assert.Equal(path, "")
		assert.Error(err)
	})
}

func TestNormalizePath(t *testing.T) {
	assert := assert.New(t)
	t.Run("test with backslash path", func(t *testing.T) {
		path := NormalizePath("test\\input\\test.csolution.yml")
		assert.Equal(path, "test/input/test.csolution.yml")
	})

	t.Run("test NormalizePath", func(t *testing.T) {
		path := NormalizePath("test/input/test.csolution.yml")
		assert.Equal(path, "test/input/test.csolution.yml")
	})
}

func TestResolveContexts(t *testing.T) {
	assert := assert.New(t)

	allContexts := []string{
		"Project1.Debug+Target",
		"Project1.Release+Target",
		"Project1.Debug+Target2",
		"Project1.Release+Target2",
		"Project2.Debug+Target",
		"Project2.Release+Target",
		"Project2.Debug+Target2",
		"Project2.Release+Target2",
		"Project3.Debug",
		"Project4+Target",
	}

	testCases := []struct {
		contextFilters           []string
		expectedResolvedContexts []string
		ExpectError              bool
	}{
		{[]string{"Project1"}, []string{"Project1.Debug+Target", "Project1.Release+Target", "Project1.Debug+Target2", "Project1.Release+Target2"}, false},
		{[]string{".Debug"}, []string{"Project1.Debug+Target", "Project1.Debug+Target2", "Project2.Debug+Target", "Project2.Debug+Target2", "Project3.Debug"}, false},
		{[]string{"+Target"}, []string{"Project1.Debug+Target", "Project1.Release+Target", "Project2.Debug+Target", "Project2.Release+Target", "Project4+Target"}, false},
		{[]string{"Project1.Debug"}, []string{"Project1.Debug+Target", "Project1.Debug+Target2"}, false},
		{[]string{"Project1+Target"}, []string{"Project1.Debug+Target", "Project1.Release+Target"}, false},
		{[]string{".Release+Target2"}, []string{"Project1.Release+Target2", "Project2.Release+Target2"}, false},
		{[]string{"Project1.Release+Target2"}, []string{"Project1.Release+Target2"}, false},

		{[]string{"*"}, allContexts, false},
		{[]string{"*.*+*"}, allContexts, false},
		{[]string{"*.*"}, allContexts, false},
		{[]string{"Proj*"}, allContexts, false},
		{[]string{".De*"}, []string{"Project1.Debug+Target", "Project1.Debug+Target2", "Project2.Debug+Target", "Project2.Debug+Target2", "Project3.Debug"}, false},
		{[]string{"+Tar*"}, []string{"Project1.Debug+Target", "Project1.Release+Target", "Project1.Debug+Target2", "Project1.Release+Target2", "Project2.Debug+Target", "Project2.Release+Target", "Project2.Debug+Target2", "Project2.Release+Target2", "Project4+Target"}, false},
		{[]string{"Proj*.D*g"}, []string{"Project1.Debug+Target", "Project1.Debug+Target2", "Project2.Debug+Target", "Project2.Debug+Target2", "Project3.Debug"}, false},
		{[]string{"Proj*+Tar*"}, []string{"Project1.Debug+Target", "Project1.Release+Target", "Project1.Debug+Target2", "Project1.Release+Target2", "Project2.Debug+Target", "Project2.Release+Target", "Project2.Debug+Target2", "Project2.Release+Target2", "Project4+Target"}, false},
		{[]string{"Project2.Rel*+Tar*"}, []string{"Project2.Release+Target", "Project2.Release+Target2"}, false},
		{[]string{".Rel*+*2"}, []string{"Project1.Release+Target2", "Project2.Release+Target2"}, false},
		{[]string{"Project*.Release+*"}, []string{"Project1.Release+Target", "Project1.Release+Target2", "Project2.Release+Target", "Project2.Release+Target2"}, false},

		// negative tests
		{[]string{"Unknown"}, nil, true},
		{[]string{".UnknownBuild"}, nil, true},
		{[]string{"+UnknownTarget"}, nil, true},
		{[]string{"Project.UnknownBuild"}, nil, true},
		{[]string{"Project+UnknownTarget"}, nil, true},
		{[]string{".UnknownBuild+Target"}, nil, true},
		{[]string{"+Debug"}, nil, true},
		{[]string{".Target"}, nil, true},
		{[]string{"TestProject*"}, nil, true},
		{[]string{"Project.*Build"}, nil, true},
		{[]string{"Project.Debug+*H"}, nil, true},
		{[]string{"Project1.Release.Debug+Target"}, nil, true},
		{[]string{"Project1.Debug+Target+Target2"}, nil, true},
	}

	for _, test := range testCases {
		outResolvedContexts, err := ResolveContexts(allContexts, test.contextFilters)
		if test.ExpectError {
			assert.Error(err)
		} else {
			assert.Nil(err)
		}
		assert.Equal(test.expectedResolvedContexts, outResolvedContexts)
	}
}

func TestRemoveDuplicates(t *testing.T) {
	assert := assert.New(t)

	inputList := []string{"apple", "banana", "apple", "orange", "banana", "grape"}
	UniqueList := []string{"apple", "banana", "orange", "grape"}

	outUniqueList := RemoveDuplicates(inputList)
	assert.Equal(UniqueList, outUniqueList)

	outUniqueList = RemoveDuplicates(UniqueList)
	assert.Equal(UniqueList, outUniqueList)
}

func TestFileExists(t *testing.T) {
	testFile := testRoot + "/" + testDir + "/testfile.txt"

	tests := []struct {
		name         string
		filePath     string
		expectedBool bool
		expectedErr  error
	}{
		{
			name:         "Existing File",
			filePath:     testFile,
			expectedBool: true,
			expectedErr:  nil,
		},
		{
			name:         "Non-Existing File",
			filePath:     "testdata/nonexistent.txt",
			expectedBool: false,
			expectedErr:  errutils.New(errutils.ErrFileNotExist, "testdata/nonexistent.txt"),
		},
		{
			name:         "Invalid Path",
			filePath:     "/invalid/path/here",
			expectedBool: false,
			expectedErr:  errutils.New(errutils.ErrFileNotExist, "/invalid/path/here"),
		},
	}

	// Create test files
	createTestFiles := func(testFile string) {
		// Create a dummy test file
		file, _ := os.Create(testFile)
		file.Close()
	}

	removeTestFiles := func(testFile string) {
		// Remove dummy test file
		os.Remove(testFile)
	}

	createTestFiles(testFile)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := FileExists(tt.filePath)
			if exists != tt.expectedBool {
				t.Errorf("Expected file existence %v, got %v", tt.expectedBool, exists)
			}
			if (err == nil && tt.expectedErr != nil) || (err != nil && err.Error() != tt.expectedErr.Error()) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}

	// Clean up test files
	removeTestFiles(testFile)
}

func TestComparePaths(t *testing.T) {
	t.Run("Windows", func(t *testing.T) {
		if strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
			// Windows-like paths (case-insensitive)
			path1 := "C:\\Users\\Example"
			path2 := "C:/users/example"
			equal, err := ComparePaths(path1, path2)
			assert.NoError(t, err, "Error should be nil")
			assert.True(t, equal, "Paths should be considered equivalent")
		}
	})

	t.Run("Darwin", func(t *testing.T) {
		if strings.Contains(strings.ToLower(os.Getenv("OS")), "darwin") {
			// macOS paths (case-insensitive)
			path1 := "/usr/Local/bin/Test"
			path2 := "/Usr/local/biN/tesT"
			equal, err := ComparePaths(path1, path2)
			assert.NoError(t, err, "Error should be nil")
			assert.True(t, equal, "Paths should be considered equivalent")
		}
	})

	t.Run("Linux", func(t *testing.T) {
		if strings.Contains(strings.ToLower(os.Getenv("OS")), "linux") {
			// Linux-like paths (case-sensitive)
			path1 := "/home/user/file.txt"
			path2 := "/home/User/FILE.txt"
			equal, err := ComparePaths(path1, path2)
			assert.NoError(t, err, "Error should be nil")
			assert.False(t, equal, "Paths should not be considered equivalent")
		}
	})
}

func TestGetTmpDir(t *testing.T) {
	t.Run("File exists with specified tmpdir", func(t *testing.T) {
		csolutionFile := filepath.Join(testRoot, testDir, "TestSolution/test.csolution.yml")
		tmpDir, err := GetTmpDir(csolutionFile, "")

		assert.NoError(t, err)
		assert.Equal(t, NormalizePath(filepath.Join(filepath.Dir(csolutionFile), "tmpdir")), tmpDir)
	})

	t.Run("File does not exist", func(t *testing.T) {
		csolutionFile := filepath.Join(testRoot, testDir, "TestSolution/non_existing.csolution.yml")
		tmpDir, err := GetTmpDir(csolutionFile, "")

		assert.ErrorIs(t, err, os.ErrNotExist)
		assert.Equal(t, "", tmpDir)
	})
}

func TestGetOutDir(t *testing.T) {
	t.Run("Index file does not exist", func(t *testing.T) {
		cbuildIdxFile := filepath.Join(testRoot, testDir, "TestSolution/non_existing.csolution.yml")
		defaultOutPath := filepath.Join(filepath.Dir(cbuildIdxFile), "out")
		outDir, err := GetOutDir(cbuildIdxFile, "test1.Debug+CM0")

		assert.NoError(t, err)
		assert.Equal(t, defaultOutPath, outDir)
	})

	t.Run("Context not found in index file", func(t *testing.T) {
		cbuildIdxFile := filepath.Join(testRoot, testDir, "TestSolution/test.cbuild-idx.yml")
		defaultOutPath := filepath.Join(filepath.Dir(cbuildIdxFile), "out")

		outDir, err := GetOutDir(cbuildIdxFile, "NonexistentContext.Debug+CM0")
		assert.NoError(t, err)
		assert.Equal(t, defaultOutPath, outDir)
	})

	t.Run("Cbuild file does not exist", func(t *testing.T) {
		cbuildIdxFile := filepath.Join(testRoot, testDir, "TestSolution/test.cbuild-idx.yml")

		outDir, err := GetOutDir(cbuildIdxFile, "test2.Debug+CM0")
		assert.Error(t, err)
		assert.Empty(t, outDir)
	})
}

func TestDeleteAll(t *testing.T) {
	testDir := filepath.Join(testRoot, testDir)
	t.Run("Delete Existing Directory", func(t *testing.T) {
		// Create a test directory with files and subdirectories
		delDir := filepath.Join(testDir, "test_dir")
		subDir := filepath.Join(delDir, "sub_dir")
		filePath := filepath.Join(delDir, "test_file.txt")
		_ = os.MkdirAll(subDir, 0755)
		_ = os.WriteFile(filePath, []byte("test content"), 0600)

		// Ensure directory exists before deletion
		if _, err := os.Stat(delDir); os.IsNotExist(err) {
			t.Fatalf("Test setup failed: %s does not exist", delDir)
		}

		// Call the DeleteAll function
		err := DeleteAll(delDir, []string{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify directory has been deleted
		if _, err := os.Stat(delDir); !os.IsNotExist(err) {
			t.Fatalf("Directory was not deleted: %s", delDir)
		}
	})

	t.Run("Delete NonExistent Directory", func(t *testing.T) {
		// Test deleting a non-existent directory
		nonExistentDir := filepath.Join(testDir, "non_existent_dir")
		err := DeleteAll(nonExistentDir, []string{})

		// Verify no error
		assert.Error(t, err)
	})

	t.Run("Delete Empty Directory", func(t *testing.T) {
		// Create an empty test directory
		emptyDir := filepath.Join(testDir, "empty_dir")
		_ = os.Mkdir(emptyDir, 0755)

		// Ensure directory exists before deletion
		if _, err := os.Stat(emptyDir); os.IsNotExist(err) {
			t.Fatalf("Test setup failed: %s does not exist", emptyDir)
		}

		// Call the DeleteAll function
		err := DeleteAll(emptyDir, []string{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify directory has been deleted
		if _, err := os.Stat(emptyDir); !os.IsNotExist(err) {
			t.Fatalf("Directory was not deleted: %s", emptyDir)
		}
	})

	t.Run("Delete File Instead Of Directory", func(t *testing.T) {
		// Create a test file
		testFile := filepath.Join(testDir, "test_file.txt")
		_ = os.WriteFile(testFile, []byte("test content"), 0600)

		// Ensure file exists before deletion
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Fatalf("Test setup failed: %s does not exist", testFile)
		}

		// Call the DeleteAll function
		err := DeleteAll(testFile, []string{})
		assert.NoError(t, err)

		// Clean up
		os.Remove(testFile)
	})

	t.Run("exclude pattern for a file", func(t *testing.T) {
		delDir := filepath.Join(testDir, "test_dir_exclude_pattern")
		_ = os.MkdirAll(delDir, 0755)
		filePath := filepath.Join(delDir, "keep.log")
		_ = os.WriteFile(filePath, []byte("test content 1"), 0600)
		filePath = filepath.Join(delDir, "delete.txt")
		_ = os.WriteFile(filePath, []byte("test content 2"), 0600)

		err := DeleteAll(delDir, []string{"*.log"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if _, err := os.Stat(filepath.Join(delDir, "keep.log")); os.IsNotExist(err) {
			t.Error("excluded file was deleted")
		}
		if _, err := os.Stat(filepath.Join(delDir, "delete.txt")); !os.IsNotExist(err) {
			t.Error("non-excluded file was not deleted")
		}
	})

	t.Run("non-existent root path", func(t *testing.T) {
		err := DeleteAll("/non/existent/path", []string{})
		if err == nil {
			t.Error("expected error for non-existent path, got nil")
		}
	})

	t.Run("exclude pattern does not match anything", func(t *testing.T) {
		delDir := filepath.Join(testDir, "test_dir_pattern_not_matching")
		_ = os.MkdirAll(delDir, 0755)
		filePath := filepath.Join(delDir, "keep.log")
		_ = os.WriteFile(filePath, []byte("test content 1"), 0600)
		filePath = filepath.Join(delDir, "delete.txt")
		_ = os.WriteFile(filePath, []byte("test content 2"), 0600)

		err := DeleteAll(delDir, []string{"nonmatch/**"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		entries, _ := os.ReadDir(delDir)
		if len(entries) != 0 {
			t.Error("expected all files deleted")
		}
	})

	t.Run("exclude nested directory", func(t *testing.T) {
		delDir := filepath.Join(testDir, "test_dir_exclude_nested_directories")
		logDir := filepath.Join(delDir, "logs")
		_ = os.MkdirAll(logDir, 0755)
		dataDir := filepath.Join(delDir, "data")
		_ = os.MkdirAll(dataDir, 0755)

		logFile1 := filepath.Join(logDir, "log1.txt")
		_ = os.WriteFile(logFile1, []byte("log1 text"), 0600)
		logFile2 := filepath.Join(logDir, "log2.debug")
		_ = os.WriteFile(logFile2, []byte("log2 debug"), 0600)
		filePath := filepath.Join(dataDir, "file.info")
		_ = os.WriteFile(filePath, []byte("data"), 0600)

		err := DeleteAll(delDir, []string{"*.debug", "*.info"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if _, err := os.Stat(filepath.Join(delDir, "logs/log2.debug")); os.IsNotExist(err) {
			t.Error("excluded nested file was deleted")
		}
		if _, err := os.Stat(filepath.Join(delDir, "logs/log1.txt")); !os.IsNotExist(err) {
			t.Error("excluded nested file was not deleted")
		}
		if _, err := os.Stat(filepath.Join(delDir, "data/file.info")); os.IsNotExist(err) {
			t.Error("excluded nested file was deleted")
		}
	})
}

func TestParseAndFetchToolchainInfo(t *testing.T) {
	// Helper function to create a temporary toolchain file
	createTempFile := func(content string) string {
		tmpFile, err := os.CreateTemp("", "toolchain_*.cmake")
		if err != nil {
			t.Fatalf("Failed to create temporary file: %v", err)
		}
		defer tmpFile.Close()

		_, err = tmpFile.WriteString(content)
		if err != nil {
			t.Fatalf("Failed to write to temporary file: %v", err)
		}

		return tmpFile.Name()
	}

	// Test cases
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "Valid Toolchain File",
			content: `
set(REGISTERED_TOOLCHAIN_ROOT "C:/test/ArmCompilerforEmbedded6.22/bin")
set(REGISTERED_TOOLCHAIN_VERSION "6.22.0")
include("${CMSIS_COMPILER_ROOT}/AC6.6.12.1.cmake")
`,
			expected: "Using AC6 V6.22.0 compiler, from: 'C:/test/ArmCompilerforEmbedded6.22/bin'",
		},
		{
			name: "Missing Info",
			content: `
set(REGISTERED_TOOLCHAIN_VERSION "6.22.0")
include("${CMSIS_COMPILER_ROOT}/AC6.6.12.1.cmake")
`,
			expected: "",
		},
		{
			name: "Missing Version",
			content: `
set(REGISTERED_TOOLCHAIN_ROOT "C:/tools/ArmCompilerforEmbedded6.22/bin")
include("${CMSIS_COMPILER_ROOT}/AC6.6.12.1.cmake")
`,
			expected: "",
		},
		{
			name: "Missing Compiler Name",
			content: `
set(REGISTERED_TOOLCHAIN_ROOT "C:/tools/ArmCompilerforEmbedded6.22/bin")
set(REGISTERED_TOOLCHAIN_VERSION "6.22.0")
`,
			expected: "",
		},
		{
			name: "Different Compiler Name",
			content: `
set(REGISTERED_TOOLCHAIN_ROOT "C:/tools/GCCCompiler/bin")
set(REGISTERED_TOOLCHAIN_VERSION "10.3.1")
include("${CMSIS_COMPILER_ROOT}/GCC.10.3.1.cmake")
`,
			expected: "Using GCC V10.3.1 compiler, from: 'C:/tools/GCCCompiler/bin'",
		},
		{
			name:     "Empty File",
			content:  ``,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file with test content
			tempFile := createTempFile(tt.content)
			defer os.Remove(tempFile)

			result := ParseAndFetchToolchainInfo(tempFile)
			assert.Equal(t, result, tt.expected)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestGetParentFolder tests the GetParentFolder function
func TestGetParentFolder(t *testing.T) {
	tests := []struct {
		name      string
		inputPath string
		want      string
		wantErr   bool
	}{
		{
			name:      "Absolute Path",
			inputPath: "/home/user/docs",
			want:      "user",
			wantErr:   false,
		},
		{
			name:      "Relative Path",
			inputPath: "./testdata/subdir",
			want:      "testdata",
			wantErr:   false,
		},
		{
			name:      "Empty Path",
			inputPath: "",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "Non-Existent Path",
			inputPath: "/this/path/does/not/exist",
			want:      "not",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetParentFolder(tt.inputPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestGetParentFolderWithTempDir ensures the function works with a real temp directory.
func TestGetParentFolderWithTempDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "testdir")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	parent := filepath.Dir(tempDir)
	expected := filepath.Base(parent)

	result, err := GetParentFolder(tempDir)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
