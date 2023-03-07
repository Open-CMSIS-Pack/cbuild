/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"regexp"
	"strings"
)

func IsWildcardPattern(str string) bool {
	return strings.ContainsAny(str, "*")
}

func ToRegEx(str string) string {
	from := []string{".", "$", "+", "{", "}", "(", ")", "?", "*"}
	to := []string{"\\.", "\\$", "\\+", "\\{", "\\}", "\\(", "\\)", ".", ".*"}
	outStr := str
	for index, char := range from {
		outStr = strings.ReplaceAll(outStr, char, to[index])
	}
	return outStr
}

func MatchString(str string, pattern string) (bool, error) {
	regex := "^" + ToRegEx(pattern) + "$"
	re, err := regexp.Compile(regex)
	if err != nil {
		return false, err
	}
	return re.MatchString(str), err
}
