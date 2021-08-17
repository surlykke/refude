// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package slice

import (
	"strings"
)

func Copy(sl []string) []string {
	res := make([]string, len(sl))
	for i, s := range sl {
		res[i] = s
	}

	return res
}

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

func PushFront(s string, list []string) []string {
	res := make([]string, 1+len(list))
	res[0] = s
	for i, item := range list {
		res[i+1] = item
	}

	return res
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

func PrependEach(sl []string, prefix string) []string {
	return Map(sl, func(s string) string {
		return prefix + s
	})
}

func Uint32SliceContains(slice []uint32, val uint32) bool {
	for _, s := range slice {
		if s == val {
			return true
			// Len is the number of elements in the collection.

		}
	}
	return false
}

type SortableStringSlice []string

func (sss SortableStringSlice) Len() int               { return len(sss) }
func (sss SortableStringSlice) Less(i int, j int) bool { return sss[i] < sss[j] }
func (sss SortableStringSlice) Swap(i int, j int)      { sss[i], sss[j] = sss[j], sss[i] }
