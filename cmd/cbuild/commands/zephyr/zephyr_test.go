/*
 * Copyright (c) 2026 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package zephyr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	"github.com/spf13/cobra"
)

type fakeRunner struct {
	commands []string
	onExec   func(entry string) error
}

func (f *fakeRunner) ExecuteCommand(program string, _ bool, args ...string) (string, error) {
	entry := filepath.Base(program)
	if len(args) > 0 {
		entry += " " + strings.Join(args, " ")
	}
	f.commands = append(f.commands, entry)
	if f.onExec != nil {
		if err := f.onExec(entry); err != nil {
			return "", err
		}
	}

	if strings.Contains(entry, "list packs") {
		return "Vendor::TestPack@1.0.0", nil
	}
	return "", nil
}

func testZephyrCmd(t *testing.T, moduleName string, clayers []string, packs []string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.Flags().String("module", moduleName, "")
	cmd.Flags().String("toolchain", defaultToolchain, "")
	cmd.Flags().String("device", defaultDevice, "")
	cmd.Flags().StringSlice("clayer", clayers, "")
	cmd.Flags().StringSlice("packs", packs, "")
	return cmd
}

func createTool(t *testing.T, binDir string, name string) {
	t.Helper()
	tool := filepath.Join(binDir, name)
	if err := os.WriteFile(tool, []byte(""), 0600); err != nil {
		t.Fatalf("failed to create %s: %v", name, err)
	}
}

func withTempWorkingDir(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	workDir := filepath.Join(tempDir, "workspace", "project")
	if err := os.MkdirAll(workDir, 0700); err != nil {
		t.Fatalf("failed to create work dir: %v", err)
	}
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if err = os.Chdir(workDir); err != nil {
		t.Fatalf("failed to set cwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
	})
	return workDir
}

func TestResolveAndValidateLayers(t *testing.T) {
	tempDir := t.TempDir()

	relativeFile := filepath.Join(tempDir, "test.clayer.yml")
	if err := os.WriteFile(relativeFile, []byte("layer"), 0600); err != nil {
		t.Fatalf("failed to create test clayer file: %v", err)
	}

	absoluteFile, err := filepath.Abs(relativeFile)
	if err != nil {
		t.Fatalf("failed to resolve absolute test path: %v", err)
	}

	t.Run("resolves relative file to absolute", func(t *testing.T) {
		layers, err := resolveAndValidateLayers([]string{"test.clayer.yml"}, tempDir)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(layers) != 1 {
			t.Fatalf("expected 1 layer, got %d", len(layers))
		}
		if layers[0] != absoluteFile {
			t.Fatalf("expected %s, got %s", absoluteFile, layers[0])
		}
	})

	t.Run("accepts absolute path", func(t *testing.T) {
		layers, err := resolveAndValidateLayers([]string{absoluteFile}, tempDir)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(layers) != 1 {
			t.Fatalf("expected 1 layer, got %d", len(layers))
		}
		if layers[0] != absoluteFile {
			t.Fatalf("expected %s, got %s", absoluteFile, layers[0])
		}
	})

	t.Run("rejects missing file", func(t *testing.T) {
		_, err := resolveAndValidateLayers([]string{"missing.clayer.yml"}, tempDir)
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("rejects invalid extension", func(t *testing.T) {
		invalid := filepath.Join(tempDir, "bad.yml")
		if err := os.WriteFile(invalid, []byte("layer"), 0600); err != nil {
			t.Fatalf("failed to create invalid test file: %v", err)
		}

		_, err := resolveAndValidateLayers([]string{"bad.yml"}, tempDir)
		if err == nil {
			t.Fatal("expected error for invalid extension")
		}
	})
}

func TestMakeLayersRelativeToProject(t *testing.T) {
	tempDir := t.TempDir()
	layersDir := filepath.Join(tempDir, "layers")
	if err := os.MkdirAll(layersDir, 0700); err != nil {
		t.Fatalf("failed to create layers dir: %v", err)
	}

	clayer := filepath.Join(layersDir, "input.clayer.yml")
	if err := os.WriteFile(clayer, []byte("layer"), 0600); err != nil {
		t.Fatalf("failed to create clayer file: %v", err)
	}

	t.Run("converts absolute clayer to path relative to cproject", func(t *testing.T) {
		projectDir := filepath.Join(tempDir, "module")
		if err := os.MkdirAll(projectDir, 0700); err != nil {
			t.Fatalf("failed to create module dir: %v", err)
		}

		projectFile := filepath.Join(projectDir, "demo.cproject.yml")
		relative, err := makeLayersRelativeToProject([]string{clayer}, projectFile)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(relative) != 1 {
			t.Fatalf("expected 1 layer, got %d", len(relative))
		}

		expected := filepath.Join("..", "layers", "input.clayer.yml")
		if relative[0] != filepath.ToSlash(expected) {
			t.Fatalf("expected %s, got %s", expected, relative[0])
		}
	})

	t.Run("keeps filename when clayer in same dir as cproject", func(t *testing.T) {
		projectDir := filepath.Join(tempDir, "same")
		if err := os.MkdirAll(projectDir, 0700); err != nil {
			t.Fatalf("failed to create same dir: %v", err)
		}

		sameLayer := filepath.Join(projectDir, "same.clayer.yml")
		if err := os.WriteFile(sameLayer, []byte("layer"), 0600); err != nil {
			t.Fatalf("failed to create same-dir layer: %v", err)
		}

		projectFile := filepath.Join(projectDir, "demo.cproject.yml")
		relative, err := makeLayersRelativeToProject([]string{sameLayer}, projectFile)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(relative) != 1 {
			t.Fatalf("expected 1 layer, got %d", len(relative))
		}

		if relative[0] != "same.clayer.yml" {
			t.Fatalf("expected same.clayer.yml, got %s", relative[0])
		}
	})
}

func TestNormalizeLayers(t *testing.T) {
	layers := normalizeLayers([]string{"  a.clayer.yml ", "", "  ", "b.clayer.yml"})
	if len(layers) != 2 {
		t.Fatalf("expected 2 layers, got %d", len(layers))
	}
	if layers[0] != "a.clayer.yml" || layers[1] != "b.clayer.yml" {
		t.Fatalf("unexpected layers: %v", layers)
	}
}

func TestValidation(t *testing.T) {
	workDir := withTempWorkingDir(t)
	clayer := filepath.Join(workDir, "ok.clayer.yml")
	if err := os.WriteFile(clayer, []byte("layer"), 0600); err != nil {
		t.Fatalf("failed to write clayer: %v", err)
	}

	t.Run("requires module", func(t *testing.T) {
		cmd := testZephyrCmd(t, "", []string{clayer}, []string{defaultPack})
		if err := generateZephyrModule(cmd, nil); err == nil {
			t.Fatal("expected error when module is missing")
		}
	})

	t.Run("requires at least one clayer", func(t *testing.T) {
		cmd := testZephyrCmd(t, "demo", []string{}, []string{defaultPack})
		if err := generateZephyrModule(cmd, nil); err == nil {
			t.Fatal("expected error when clayer is missing")
		}
	})
}

func TestZephyrModuleGeneration(t *testing.T) {
	workDir := withTempWorkingDir(t)
	destinationModuleDir := filepath.Join(workDir, "demo")
	if err := os.MkdirAll(destinationModuleDir, 0700); err != nil {
		t.Fatalf("failed to create destination module dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(destinationModuleDir, "old.txt"), []byte("old"), 0600); err != nil {
		t.Fatalf("failed to create destination marker file: %v", err)
	}

	binDir := filepath.Join(workDir, "bin")
	if err := os.MkdirAll(binDir, 0700); err != nil {
		t.Fatalf("failed to create bin dir: %v", err)
	}
	createTool(t, binDir, "csolution")
	createTool(t, binDir, "cbuild2cmake")
	createTool(t, binDir, "cpackget")

	clayer := filepath.Join(workDir, "input.clayer.yml")
	if err := os.WriteFile(clayer, []byte("layer"), 0600); err != nil {
		t.Fatalf("failed to write clayer: %v", err)
	}

	oldGetInstallConfigs := getInstallConfigs
	oldRunnerInterface := runnerInterface
	t.Cleanup(func() {
		getInstallConfigs = oldGetInstallConfigs
		runnerInterface = oldRunnerInterface
	})

	getInstallConfigs = func() (utils.Configurations, error) {
		return utils.Configurations{BinPath: binDir, EtcPath: workDir, BinExtn: ""}, nil
	}
	fake := &fakeRunner{onExec: func(entry string) error {
		if strings.Contains(entry, "cbuild2cmake demo.cbuild-idx.yml --zephyr") {
			if err := os.MkdirAll("demo", 0700); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join("demo", "new.txt"), []byte("new"), 0600); err != nil {
				return err
			}
		}
		return nil
	}}
	runnerInterface = func() utils.RunnerInterface { return fake }

	cmd := testZephyrCmd(t, "demo", []string{clayer}, []string{defaultPack})
	if err := generateZephyrModule(cmd, nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	joined := strings.Join(fake.commands, "\n")
	listIndex := strings.Index(joined, "csolution list packs")
	addIndex := strings.Index(joined, "cpackget add Vendor::TestPack@1.0.0")
	convertIndex := strings.Index(joined, "csolution demo.csolution.yml convert")
	cmakeIndex := strings.Index(joined, "cbuild2cmake demo.cbuild-idx.yml --zephyr")

	if listIndex == -1 || addIndex == -1 || convertIndex == -1 || cmakeIndex == -1 {
		t.Fatalf("unexpected command sequence:\n%s", joined)
	}
	if listIndex >= addIndex || addIndex >= convertIndex || convertIndex >= cmakeIndex {
		t.Fatalf("commands are not in expected order:\n%s", joined)
	}
	if _, err := os.Stat(filepath.Join(destinationModuleDir, "new.txt")); err != nil {
		t.Fatalf("expected generated module to be moved to destination, got: %v", err)
	}
	if _, err := os.Stat(filepath.Join(destinationModuleDir, "old.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected old destination content to be replaced, got err: %v", err)
	}
}
