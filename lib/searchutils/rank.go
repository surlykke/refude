package searchutils

import (
	"strings"

	"github.com/surlykke/RefudeServices/lib/respond"
)

// Kindof 'has Substring with skips'. So eg. 'nvim' matches 'neovim' or 'pwr' matches 'poweroff'
// term must not be empty
func FluffyRank(text string, term []rune) int {
	var j = 0
	for i, r := range text {
		if term[j] == r {
			j++
		}
		if j >= len(term) {
			return 1000 + i // FIXME return a rank that expresses the 'spread' of the match
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
			return rank
		}
	}

	return -1
}
