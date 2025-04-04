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
	ErrInvalidFileExtension   = "invalid file extension: '%s'. Expected: '%s'"
	ErrInvalidCmdLineArg      = "invalid command line argument"
	ErrFileNotExist           = "file %s does not exist"
	ErrNoContextFound         = "no context found to process"
	ErrBinaryNotFound         = "%s not found %s"
	ErrMissingPacks           = "missing packs. Use --packs option with cbuild command to install them"
	ErrETCPathNotFound        = "couldn't locate '%s' directory relative to '%s'"
	ErrInvalidContextFormat   = "invalid context format. Expected [<project-name>][.<build-type>][+<target-type>]"
	ErrNoFilteredContextFound = "no valid context found for '%s'"
	ErrInvalidCommand         = "invalid command '%s'. Run 'cbuild --help' for supported commands"
	ErrCPRJNotFound           = "couldn't locate %s file"
	ErrNinjaVersionNotFound   = "unable to find 'ninja' version"
	ErrAcceptNoArgs           = "command does not accept any arguments. Run '%s' for more information about a command"
	ErrRequireArg             = "command requires an input file argument. Run '%s' for more information about a command"
	ErrInvalidVersionString   = "invalid version %s. Expected %s"
	ErrInvalidNumJobs         = "invalid number of job slots specified for parallel execution. Expected: j>0"
	ErrMissingRequiredArg     = "setup command is missing mandatory option '--context-set'"
	ErrDeleteFailed           = "failed to delete: '%s'"
	ErrPathNotExist           = "path does not exist: '%s'"
	ErrFetchingAbsPath        = "unable to get absolute path: '%s'"
	ErrInvalidPath            = "invalid path: '%s'"
	ErrPerfResults            = "unable to save performance results: %s"
	ErrNoCompilerRegistered   = "required compiler(s) not registered: '%s'"
)

const (
	WarnNinjaVersion = "use Ninja 1.11.1 or higher for less verbose output"
)

func New(errorFormat string, args ...any) error {
	errMsg := fmt.Sprintf(errorFormat, args...)
	return errors.New(errMsg)
}
