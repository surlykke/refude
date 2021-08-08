package searchutils

import (
	"strings"
)

// Kindof 'index with skips'. So eg. 'nvim' wille be found in 'neovim' or 'pwr' in 'poweroff'
// case sensitive, -1 means not found
func FluffyIndex(haystack, needle []rune) int {
	if len(haystack) == 0 {
		return -1
	}
	// Special handling targeted at at hidden files. Those are the only names (I think)
	// starting with '.'. We only show them if explicitly sought for (term starts with .)
	if haystack[0] == '.' && (len(needle) == 0 || needle[0] != '.') {
		return -1
	}

	var j = -1
	var first = -1
	for _, needleRune := range needle {
		for {
			j++
			if j >= len(haystack) {
				return -1
			} else if haystack[j] == needleRune {
				if first == -1 {
					first = j
				}
				break
			}
		}
	}
	var rnk = first + 5*((j+1-first)-len(needle)) // kindof start of match + spread of match
	return rnk
}

func Match(term, name string, keywords ...string) int {
	term = strings.ToLower(term)
	name = strings.ToLower(name)
	if rnk := FluffyIndex([]rune(name), []rune(term)); rnk > -1 {
		return rnk
	}
	for _, keyword := range keywords {
		if strings.Index(strings.ToLower(keyword), term) > 0 {
			return 1000
		}
	}
	return -1
}
