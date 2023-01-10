/*
 * Copyright (c) 2022-2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cproject

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

type Cprj struct {
	XMLName      xml.Name     `xml:"cprj"`
	TargetOutput TargetOutput `xml:"target>output"`
}

type TargetOutput struct {
	XMLName xml.Name `xml:"output"`
	IntDir  string   `xml:"intdir,attr"`
	OutDir  string   `xml:"outdir,attr"`
}

func GetCprjDirs(file string) (string, string, error) {
	xmlFile, err := os.Open(file)
	if err != nil {
		return "", "", err
	}
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	var cprj Cprj
	err = xml.Unmarshal(byteValue, &cprj)
	if err != nil {
		return "", "", err
	}

	return cprj.TargetOutput.IntDir, cprj.TargetOutput.OutDir, err
}
