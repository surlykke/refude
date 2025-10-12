package search

import "testing"

func TestMatch(t *testing.T) {
	var matcher = makeMatcher("abc")
	var text = "fffababcggg"
	matcher.match(text)
}
