/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cbuildidx

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Cbuild struct {
	Build struct {
		OutputDirs struct {
			Intdir string `yaml:"intdir"`
			Outdir string `yaml:"outdir"`
		} `yaml:"output-dirs"`
	} `yaml:"build"`
}

func GetBuildDirs(file string) (string, string, error) {
	yfile, err := os.ReadFile(file)
	if err != nil {
		return "", "", err
	}
	data := Cbuild{}
	err = yaml.Unmarshal(yfile, &data)
	return data.Build.OutputDirs.Intdir, data.Build.OutputDirs.Outdir, err
}
