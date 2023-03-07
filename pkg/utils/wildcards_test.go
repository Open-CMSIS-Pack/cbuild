/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsWildcardPattern(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(IsWildcardPattern(".+^$()?[]{}/.)"), false)
	assert.Equal(IsWildcardPattern("*+Target1"), true)
	assert.Equal(IsWildcardPattern(".Debug+*"), true)
	assert.Equal(IsWildcardPattern(".D*g"), true)

}

func TestToRegEx(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(ToRegEx(".Debug+*"), "\\.Debug\\+.*")
	assert.Equal(ToRegEx("+target"), "\\+target")
	assert.Equal(ToRegEx("*D*g"), ".*D.*g")
}

func TestMatchString(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		InputString   string
		InputPattern  string
		ExpectedMatch bool
		ExpectError   bool
	}{
		{"a", "a", true, false},
		{"a", "", false, false},
		{"", "d", false, false},
		{"", "*", true, false},
		{"abcd", "a*d", true, false},
		{"adb", "*d", false, false},
		{"abcd", "abcd", true, false},
		{"abcd", "xycd", false, false},
		{"add", "*d", true, false},
		{"abcd", "a*d", true, false},
		{"abxx", "a*d", false, false},
		{"abxyz", "a*d", false, false},
		{"xycd", "a*d", false, false},
		{"abcd", "*d", true, false},
		{"d", "*d", true, false},
		{"abcd", "*c*", true, false},
		{"abcd", "a**d", true, false},
		{"abcd", "a*d", true, false},
		{"abcd", "*bc*", true, false},
		{"abcd", "abc*", true, false},
		{"abcd", "ab*", true, false},
		{"abcX-1", "abcX-2", false, false},
		{"abcX-1", "abcX-3", false, false},
		{"abcX-1", "abcY-1", false, false},
		{"abcX-1", "abcY-2", false, false},

		{"Prefix_Mid_Suffix", "Prefix_*_Suffix", true, false},
		{"Prefix_Mid_V_Suffix", "Prefix_*_Suffix", true, false},
		{"Prefix_Mid_Suffix_Suffix", "Prefix_*_Suffix", true, false},
		{"Prefix_Mid_Suffix", "Prefix*_Suffix", true, false},
		{"Prefix_Mid_Suffix", "Prefix*Suffix", true, false},
		{"Prefix_Mid_Suffix_Suffix", "Prefix*Suffix", true, false},
		{"Prefix_Mid_Suffix", "Prefix_*Suffix", true, false},
		{"Prefix.Mid.Suffix", "Prefix.*.Suffix", true, false},
		{"Prefix.Mid+Suffix", "Prefix.*+Suffix", true, false},

		{"Prefix_${Mid_}Suffix", "Prefix_${*}Suffix", true, false},
		{"Prefix_$(Mid_)Suffix", "Prefix_$(*)Suffix", true, false},
		{"Prefix_\\(Suffix", "Prefix_\\(*Suffix", false, true},
	}

	for _, test := range testCases {
		match, err := MatchString(test.InputString, test.InputPattern)
		assert.Equal(match, test.ExpectedMatch, "failed for input '"+test.InputString+"' '"+test.InputPattern+"'")
		if test.ExpectError {
			assert.Error(err)
		} else {
			assert.Nil(err)
		}
	}
}
