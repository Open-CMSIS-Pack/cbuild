/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package errutils

import (
	"errors"
	"fmt"
)

const (
	ErrInvalidFileExtension     = "unsupported file extension: %s. Supported extension(s): %s"
	ErrInvalidCmdLineArg        = "expected only one argument specifying the input file"
	ErrFileNotExist             = "file %s does not exist"
	ErrNoContextFound           = "no context(s) found to process"
	ErrBinaryNotFound           = "%s binary not found %s"
	ErrMissingPacks             = "missing packs must be installed, rerun cbuild with the --packs option"
	ErrPathNotFound             = "path %s not found"
	ErrInvalidContextFormat     = "invalid context format. Expected [project][.buildType][+targetType]"
	ErrInvalidCSolutionFileName = "invalid csolution file name format. Expected '<projectname>.csolution.yml'"
	ErrNoFilteredContextFound   = "no suitable context matched for filter '%s'"
	ErrNoRefToCPRJFile          = "reference to '%s' not found in '%s' file"
	ErrInvalidCommand           = "invalid command entered. Please check your input and try again"
	ErrInvalidFile              = "invalid file: %s. Expected '%s' file"
)

func New(errorFormat string, args ...any) error {
	errMsg := fmt.Sprintf(errorFormat, args...)
	return errors.New(errMsg)
}
