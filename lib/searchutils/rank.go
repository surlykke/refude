package searchutils

import (
	"strings"
)

// Kindof 'has Substring with skips'. So eg. 'nvim' matches 'neovim' or 'pwr' matches 'poweroff'
// term must not be empty
func fluffyRank(text string, term string) (uint, bool) {
	var runes = []rune(term)
	var j = 0
	var first = -1
	for i, r := range text {
		if runes[j] == r {
			if first == -1 {
				first = i
			}
			j++
		}
		if j >= len(term) {
			return uint(first + 5*((i+1-first)-len(term))), true // A rank that expresses the 'spread' of the match
		}
	}

	return 0, false
}

/**
 *
 * Caller must ensure term is lowercase
 */
func SimpleRank(title, comment, term string) (uint, bool) {
	if term == "" {
		return 0, true
	}
	title = strings.ToLower(title)
	var tmp = strings.Index(title, term)
	if tmp > -1 {
		return uint(tmp), true
	}
	if comment != "" {
		comment = strings.ToLower(comment)
		tmp = strings.Index(comment, term)
		if tmp > -1 {
			return uint(tmp) + 100, true
		}
	}
	if rank, ok := fluffyRank(title, term); ok {
		return 200 + rank, true
	} else {
		return 0, false
	}
}
