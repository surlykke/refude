package searchutils

import (
	"strings"
)

func Match(term, name string, keywords ...string) int {
	term = strings.ToLower(term)
	name = strings.ToLower(name)
	if rnk := strings.Index(name, term); rnk > -1 {
		return rnk
	}
	for _, keyword := range keywords {
		if strings.Index(strings.ToLower(keyword), term) > 0 {
			return 1000
		}
	}
	return -1
}
