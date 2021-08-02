package searchutils

import "github.com/surlykke/RefudeServices/lib/respond"

// Kindof 'index with skips'. So eg. 'nvim' wille be found in 'neovim' or 'pwr' in 'poweroff'
// case sensitive, -1 means not found
func FluffyIndex(haystack, needle []rune) int {
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
	return first + 5*((j+1-first)-len(needle)) // kindof start of match + spread of match
}

type Crawler func(related respond.Link, keywords []string)

type Crawl func(term string, forDisplay bool, crawler Crawler)
