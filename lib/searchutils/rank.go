package searchutils

import (
	"strings"
)

// Kindof 'has Substring with skips'. So eg. 'nvim' matches 'neovim' or 'pwr' matches 'poweroff'
// term must not be empty
func FluffyRank(text string, term []rune) int {
	var j = 0
	var first = 0
	for i, r := range text {
		if term[j] == r {
			if j == 0 {
				first = j
			}
			j++
		}
		if j >= len(term) {
			return first + 5*((i-first)-j+1) // A rank that expresses the 'spread' of the match
		}
	}

	return -1
}

// Caller ensures term is lowercase
func SimpleRank(title, comment, term string) int {
	if term == "" {
		return 0
	}
	title = strings.ToLower(title)
	var tmp = strings.Index(title, term)
	if tmp > -1 {
		return tmp
	}
	if comment != "" {
		comment = strings.ToLower(comment)
		tmp = strings.Index(comment, term)
		if tmp > -1 {
			return tmp + 100
		}
	}
	return -1
}
