// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package slice

import (
	"strings"
)

func Contains(sl []string, s ...string) bool {
	for _, str := range sl {
		for _, s2 := range s {
			if str == s2 {
				return true
			}
		}

	}

	return false
}

func Among(s string, values ...string) bool {
	for _, v := range values {
		if s == v {
			return true
		}
	}

	return false
}

func ElementsInCommon(l1 []string, l2 []string) bool {
	for _, s := range l1 {
		if Contains(l2, s) {
			return true
		}
	}

	return false
}

func AppendIfNotThere(list []string, otherList ...string) []string {
	for _, other := range otherList {
		var found = false
		for _, v := range list {
			if v == other {
				found = true
				break
			}
		}
		if !found {
			list = append(list, other)
		}
	}
	return list
}

func Remove(list []string, otherList ...string) []string {
	var pos = 0
	for i := 0; i < len(list); i++ {
		var found = false
		for _, other := range otherList {
			if other == list[i] {
				found = true
				break
			}
		}
		if !found {
			list[pos] = list[i]
			pos += 1
		}
	}
	return list[0:pos]
}

func Filter(list []string, test func(s string) bool) []string {
	res := make([]string, 0, len(list))
	for _, s := range list {
		if test(s) {
			res = append(res, s)
		}
	}

	return res
}

func Map(list []string, mapper func(s string) string) []string {
	res := make([]string, len(list))
	for i, s := range list {
		res[i] = mapper(s)
	}
	return res
}

func Split(str string, sep string) []string {
	return Filter(Map(strings.Split(str, sep),
		func(s string) string {
			return strings.TrimSpace(s)
		}),
		func(s string) bool {
			return s != ""
		})
}
