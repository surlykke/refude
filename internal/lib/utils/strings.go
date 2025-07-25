package utils

import (
	"strings"
)

/**
* Like utils.Split, but trims tokens and removes ""
 */
func Split(s, sep string) []string {
	var tokens = strings.Split(s, sep)
	var pos = 0
	for _, t := range tokens {
		t = strings.TrimSpace(t)
		if t != "" {
			tokens[pos] = t
			pos = pos + 1
		}
	}
	return tokens[0:pos]
}
