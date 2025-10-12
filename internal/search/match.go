// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package search

import (
	"strings"
)

/* Given a text (could be a resource title) and a term (what the user types to search for),
* we look for shortest possible substrings containing the runes of term in the order the appear in term.
* So if we search for 'abc'
*
*    in                  it should yield   at position
*    'a very big cat'    'a very big c'    0
*    'abcabcabc'         'abc'             0, 3 and 6
*    'aaaaabbbbcccc'     'abbbbc'          4
*    'fgabhiabjkclmn'    'abjkc'           6
*
* We can do that with a regular expression like `a[^a]* b[^c]* c`, where the terms `[^a]* ` and `[^c]'` will
* make the matched subtring start at the latest possible point and  at the earliest possible point.
*
* If we search for a longer term, eg. 'abcd' the regexp would be: `a[^a]* b.* c[^d]* d`
* If we search for a term of length 2, eg. 'ab', it's: `a[^a]* [^b]* b`
*
* We do not use regexp for this, but rather a function coded specifically for these types of regular expressions.
*
* function findMatches implements that. Given a term of length, say 4, eg 'abcd' it maintains an array, 'state' of length 3 of integers.
*
* If, say, state[1] == 7, that means that we have a potential match starting at textpos 7, and have so far read 'a' and 'b'. If we then encounter
* 'c' we set state[1] = -1 and state[2] = 7.
*
*  As another example lets search for 'abc' in 'fffababcggg'. As we read through 'fffababcggg', state will evolve as:
* Text position         letter read    match   state
*             0                   f            [-1, -1]
*             1                   f            [-1, -1]
*             2                   f            [-1, -1]
*             3                   a            [ 3, -1]
*             4                   b            [-1,  3]
*             5                   a            [ 5,  3]
*             6                   b            [-1,  5]
*             7                   c    [5:8]   [-1, -1]
*             8                   g            [-1, -1]
*             9                   g            [-1, -1]
*            10                   g            [-1, -1]
*
 */

type matcher struct {
	term      []rune
	curstate  []int
	nextstate []int
}

func makeMatcher(term string) matcher {
	var m = matcher{term: []rune(strings.ToLower(term))}
	if len(term) > 1 {
		m.curstate = make([]int, len(m.term)-1)
		m.nextstate = make([]int, len(m.term)-1)
	}
	return m
}

func (this matcher) match(text string) uint {
	text = strings.ToLower(text)
	switch len(this.term) {
	case 0:
		return 0
	case 1:
		for pos, r := range text {
			if this.term[0] == r {
				return uint(pos)
			}
		}
		return maxRank
	default:
		var res = maxRank

		for i := range this.curstate {
			this.curstate[i] = -1
		}

		for textpos, r := range text {
			if r == this.term[0] {
				this.nextstate[0] = textpos
			} else {
				this.nextstate[0] = this.curstate[0]
			}
			for i := 1; i < len(this.term)-1; i++ {
				if this.term[i] == r && this.curstate[i-1] > -1 {
					this.nextstate[i], this.nextstate[i-1] = this.curstate[i-1], -1
				} else {
					this.nextstate[i] = this.curstate[i]
				}
			}
			if this.term[len(this.term)-1] == r && this.curstate[len(this.term)-2] > -1 {
				var start, end = this.curstate[len(this.term)-2], textpos + 1
				var tmp = uint(start + 5*(end-start-len(this.term)))
				if tmp < res {
					res = tmp
				}
				this.nextstate[len(this.term)-2] = -1
			}
			this.curstate, this.nextstate = this.nextstate, this.curstate
		}
		return res
	}
}
