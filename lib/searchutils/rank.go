// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package searchutils

import (
	"github.com/lithammer/fuzzysearch/fuzzy"
)

func Match(term, name string, keywords ...string) int {
	var val = -1
	for i, target := range append([]string{name}, keywords...) {
		var dist = fuzzy.RankMatchNormalizedFold(term, target)
		if dist > -1 && i > 0 {
			dist = dist + 100
		}
		if val == -1 || (dist > -1 && dist < val) {
			val = dist
		}
	}
	return val
}
