/*
 * Copyright (c) 2022 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package commands

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"runtime"
	"testing"
	"time"

	cp "github.com/otiai10/copy"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../../test"

type RunnerMock struct{}

func (r RunnerMock) ExecuteCommand(program string, quiet bool, args ...string) ([]byte, error) {
	return nil, nil
}

func init() {
	// Prepare test data
	_ = os.RemoveAll(testRoot + "/run")
	time.Sleep(time.Second)
	_ = cp.Copy(testRoot+"/data", testRoot+"/run")

	_ = os.MkdirAll(testRoot+"/run/bin", 0755)
	_ = os.MkdirAll(testRoot+"/run/etc", 0755)
	_ = os.MkdirAll(testRoot+"/run/packs", 0755)
	_ = os.MkdirAll(testRoot+"/run/IntDir", 0755)
	_ = os.MkdirAll(testRoot+"/run/OutDir", 0755)

	var binExtension string
	if runtime.GOOS == "windows" {
		binExtension = ".exe"
	}
	cbuildgenBin := testRoot + "/run/bin/cbuildgen" + binExtension
	file, _ := os.Create(cbuildgenBin)
	defer file.Close()
	csolutionBin := testRoot + "/run/bin/csolution" + binExtension
	file, _ = os.Create(csolutionBin)
	defer file.Close()
	cpackgetBin := testRoot + "/run/bin/cpackget" + binExtension
	file, _ = os.Create(cpackgetBin)
	defer file.Close()

	_ = cp.Copy(testRoot+"/run/test.Debug+CM0.cprj", testRoot+"/run/OutDir/test.Debug+CM0.cprj")
}

func TestCommands(t *testing.T) {
	assert := assert.New(t)
	os.Setenv("CMSIS_BUILD_ROOT", testRoot+"/run/bin")
	cprjFile := testRoot + "/run/minimal.cprj"
	csolutionFile := testRoot + "/run/TestSolution/test.csolution.yml"

	t.Run("test version", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{"--version"})
		err := cmd.Execute()
		assert.Nil(err)
	})

	t.Run("test help", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{"--help"})
		err := cmd.Execute()
		assert.Nil(err)
	})

	t.Run("invalid flag", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{"--invalid"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("multiple arguments", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{cprjFile, cprjFile})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test CPRJ build", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{cprjFile})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test CSolution build", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{csolutionFile})
		err := cmd.Execute()
		assert.Error(err)
	})
}

// ====================================================
type TestCase struct {
	name           string
	args           []string
	expectedStdout []string
	expectedStderr []string
	expectedErr    error
	setUpFunc      func(t *TestCase)
	tearDownFunc   func()
}

func runTests(t *testing.T, tests []TestCase) {
	assert := assert.New(t)

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			if test.setUpFunc != nil {
				test.setUpFunc(&test)
			}

			if test.tearDownFunc != nil {
				defer test.tearDownFunc()
			}

			cmd := NewRootCmd()

			stdout := bytes.NewBufferString("")
			stderr := bytes.NewBufferString("")

			cmd.SetOut(stdout)
			cmd.SetErr(stderr)
			cmd.SetArgs(test.args)

			cmdErr := cmd.Execute()
			// Very important: resets all flags, as apparently
			// Cobra doesn't do that.
			// Otherwise, the first time a command uses a flag,
			// it will taint the others.
			// Ref: https://github.com/spf13/cobra/issues/1488
			for _, c := range cmd.Commands() {
				c.Flags().VisitAll(func(f *pflag.Flag) {
					if f.Changed {
						_ = f.Value.Set(f.DefValue)
						f.Changed = false
					}
				})
			}

			outBytes, err1 := ioutil.ReadAll(stdout)
			errBytes, err2 := ioutil.ReadAll(stderr)
			assert.Nil(err1)
			assert.Nil(err2)

			outStr := string(outBytes)
			errStr := string(errBytes)

			assert.Equal(test.expectedErr, cmdErr)
			for _, expectedStr := range test.expectedStdout {
				assert.Contains(outStr, expectedStr)
			}

			for _, expectedStr := range test.expectedStderr {
				assert.Contains(errStr, expectedStr)
			}
		})
	}
}

var rootCmdTests = []TestCase{
	{
		name:           "test get version",
		args:           []string{"--version"},
		expectedStdout: []string{"cbuild version testing#123 (C) 2022 Arm Ltd. and Contributors"},
		setUpFunc: func(t *TestCase) {
			version = "testing#123"
		},
		tearDownFunc: func() {
			version = ""
		},
	},
	{
		name:           "test help",
		args:           []string{"--help"},
		expectedStdout: []string{"Use \"cbuild [command] --help\" for more information about a command."},
		expectedErr:    nil,
	},
	{
		name:        "test unknown option",
		args:        []string{"--invalid"},
		expectedErr: errors.New("unknown flag: --invalid"),
	},
	{
		name:        "test multiple arguments",
		args:        []string{testRoot + "/run/minimal.cprj", testRoot + "/run/minimal.cprj"},
		expectedErr: errors.New("invalid arguments"),
	},
	{
		name: "test CPRJ build",
		args: []string{testRoot + "/run/minimal.cprj"},
		setUpFunc: func(t *TestCase) {
			os.Setenv("CMSIS_BUILD_ROOT", testRoot+"/run/bin")
		},
		expectedStderr: []string{"fork/exec", "Err:0xc1"},
	},
	// {
	// 	name:        "test csolution build",
	// 	args:        []string{testRoot + "/run/TestSolution/test.csolution.yml"},
	// 	expectedErr: errors.New("error cbuild: invalid arguments"),
	// },

	// {
	// 	name:           "test no parameter given",
	// 	expectedStdout: []string{"requires at least 1 arg(s), only received 0"},
	// },
}

func TestRootCmd(t *testing.T) {
	runTests(t, rootCmdTests)
}

//=======================================================

// func init() {
// 	// Prepare test data
// 	_ = os.RemoveAll(testRoot + "/run")
// 	time.Sleep(time.Second)
// 	_ = cp.Copy(testRoot+"/data", testRoot+"/run")
// }

// func runTests(t *testing.T, tests []TestCase) {
// 	assert := assert.New(t)

// 	for i := range tests {
// 		test := tests[i]
// 		t.Run(test.name, func(t *testing.T) {
// 			//localTestingDir := strings.ReplaceAll(test.name, " ", "_")
// 			//os.Setenv("CMSIS_PACK_ROOT", localTestingDir)
// 			// if test.createPackRoot {
// 			// 	assert.Nil(installer.SetPackRoot(localTestingDir, test.createPackRoot))
// 			// 	installer.UnlockPackRoot()
// 			// }

// 			// if test.env != nil {
// 			// 	for envVar := range test.env {
// 			// 		os.Setenv(envVar, test.env[envVar])
// 			// 	}
// 			// }

// 			if test.setUpFunc != nil {
// 				test.setUpFunc(&test)
// 			}

// 			if test.tearDownFunc != nil {
// 				defer test.tearDownFunc()
// 			}

// 			cmd := NewRootCmd()

// 			stdout := bytes.NewBufferString("")
// 			stderr := bytes.NewBufferString("")

// 			cmd.SetOut(stdout)
// 			cmd.SetErr(stderr)
// 			cmd.SetArgs(test.args)

// 			cmdErr := cmd.Execute()
// 			// Very important: resets all flags, as apparently
// 			// Cobra doesn't do that.
// 			// Otherwise, the first time a command uses a flag,
// 			// it will taint the others.
// 			// Ref: https://github.com/spf13/cobra/issues/1488
// 			for _, c := range cmd.Commands() {
// 				c.Flags().VisitAll(func(f *pflag.Flag) {
// 					if f.Changed {
// 						_ = f.Value.Set(f.DefValue)
// 						f.Changed = false
// 					}
// 				})
// 			}

// 			outBytes, err1 := ioutil.ReadAll(stdout)
// 			errBytes, err2 := ioutil.ReadAll(stderr)
// 			assert.Nil(err1)
// 			assert.Nil(err2)

// 			outStr := string(outBytes)
// 			errStr := string(errBytes)

// 			assert.Equal(test.expectedErr, cmdErr)
// 			for _, expectedStr := range test.expectedStdout {
// 				assert.Contains(outStr, expectedStr)
// 			}

// 			for _, expectedStr := range test.expectedStderr {
// 				assert.Contains(errStr, expectedStr)
// 			}
// 		})
// 	}
// }

// type TestCase struct {
// 	args []string
// 	name string
// 	// defaultMode    bool
// 	// createPackRoot bool
// 	expectedStdout []string
// 	expectedStderr []string
// 	expectedErr    error
// 	setUpFunc      func(t *TestCase)
// 	tearDownFunc   func()
// }

// var rootCmdTests = []TestCase{
// 	{
// 		name:           "test get version",
// 		args:           []string{"--version"},
// 		expectedStdout: []string{"cbuild version testing#123 (C) 2022 Arm Ltd. and Contributors"},
// 		setUpFunc: func(t *TestCase) {
// 			version = "testing#123"
// 		},
// 		tearDownFunc: func() {
// 			version = ""
// 		},
// 	},
// 	{
// 		name:           "test help",
// 		args:           []string{"--help"},
// 		expectedStdout: []string{"Use \"cbuild [command] --help\" for more information about a command."},
// 		expectedErr:    nil,
// 	},
// 	{
// 		name:        "test unknown option",
// 		args:        []string{"--invalid"},
// 		expectedErr: errors.New("unknown flag: --invalid"),
// 	},
// 	{
// 		name:        "test multiple arguments",
// 		args:        []string{testRoot + "/run/minimal.cprj", testRoot + "/run/minimal.cprj"},
// 		expectedErr: errors.New("unknown command \"../../../test/run/minimal.cprj\" for \"cbuild\""),
// 	},
// 	{
// 		name:        "test CPRJ build",
// 		args:        []string{testRoot + "/run/minimal.cprj"},
// 		expectedErr: errors.New("error cbuild: invalid arguments"),
// 	},
// 	// {
// 	// 	name:        "test csolution build",
// 	// 	args:        []string{testRoot + "/run/TestSolution/test.csolution.yml"},
// 	// 	expectedErr: errors.New("error cbuild: invalid arguments"),
// 	// },

// 	// {
// 	// 	name:           "test no parameter given",
// 	// 	expectedStdout: []string{"requires at least 1 arg(s), only received 0"},
// 	// },
// }

// func TestRootCmd(t *testing.T) {
// 	runTests(t, rootCmdTests)
// }
