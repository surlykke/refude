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
func FluffyRank(text, term []rune, baserank int) (int, bool) {
	var j = -1
	var first = -1
	for _, termRune := range term {
		for {
			j++
			if j >= len(text) {
				return 0, false
			} else if text[j] == termRune {
				if first == -1 {
					first = j
				}
				break
			}
		}
	}
	return baserank + first + 5*((j+1-first)-len(term)), true // A rank that expresses the 'spread' of the match
}
