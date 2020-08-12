package searchutils

import (
	"strings"
)

// case sensitive
func Rank(text, term string, baserank int) (int, bool) {
	text = strings.ToLower(text)
	if i := strings.Index(text, term); i > -1 {
		return baserank + i, true
	} else {
		return 0, false
	}
}

// Kindof 'has Substring with skips'. So eg. 'nvim' matches 'neovim' or 'pwr' matches 'poweroff'
// case sensitive
func FluffyRank(text, term string, baserank int) (int, bool) {
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
			return baserank + first + 5*((i+1-first)-len(term)), true // A rank that expresses the 'spread' of the match
		}
	}

	return 0, false
}

/**
 *
 * Caller must ensure term and comment are lowercase
 */
func SimpleRank(title, comment, term string) (int, bool) {
	if term == "" {
		return 0, true
	}
	title = strings.ToLower(title)
	if rank, ok := Rank(title, term, 0); ok {
		return rank, true
	}
	if comment != "" {
		if rank, ok := Rank(comment, term, 100); ok {
			return rank, true
		}
	}
	return FluffyRank(title, term, 200)
}
