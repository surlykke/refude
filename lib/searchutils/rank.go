package searchutils

import (
	"strings"

	"github.com/surlykke/RefudeServices/lib/respond"
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

/**
 * term and termAsRunes are assumed to represent same text, caller must ensure.
 * term and termAsRunes are assumed to be in lowerCase, caller must ensure
 */
func Rank(sf *respond.StandardFormat, term string, termAsRunes []rune) int {
	if len(term) == 0 {
		return 0
	}

	if len(sf.Title) >= len(term) {
		var title = strings.ToLower(sf.Title)
		if rank := strings.Index(title, term); rank > -1 {
			return rank
		}

		if rank := FluffyRank(title, termAsRunes); rank > -1 {
			return rank
		}

	}

	if len(sf.Comment) > len(term) {
		if rank := strings.Index(strings.ToLower(sf.Comment), term); rank > -1 {
			return rank + 1000
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
