// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
//
package search

import "testing"

func TestMatch(t *testing.T) {
	var matcher = makeMatcher("abc")
	var text = "fffababcggg"
	matcher.match(text)
}
