package main

import (
	"encoding/json"
)

type StringSet map[string]bool

func (set *StringSet) add(s string) {
	(*set)[s] = true
}

func ensureExists(m map[string]*StringSet, s string) {
	if _,ok := m[s]; !ok {
		stringSet := make(StringSet)
		m[s] = &stringSet
	}
}

func makeStringSetPtr() *StringSet {
	stringSet := make(StringSet)
	return &stringSet
}

func (set *StringSet) remove(s string) {
	delete(*set, s)
}

func (set *StringSet) addAll(list []string) {
	for _,s := range list {
		set.add(s)
	}
}

func (set *StringSet) removeAll(list []string) {
	for _,s := range list {
		set.remove(s)
	}
}

func (ss StringSet) MarshalJSON() ([]byte, error) {
    keys := make([]string, 0, len(ss))
    for key,_ := range ss {
        keys = append(keys, key)
    }
    return json.Marshal(keys)
}

func toSet(list []string) StringSet {
	res := make(StringSet)
	for _,s := range list {
		res[s] = true
	}
	return res
}

func appendIfNotThere(list []string, s string) []string {
	for _, v := range list {
		if v == s {
			return list
		}
	}

	return append(list, s)
}

func removeDublets(list []string) []string {
	seen := make(map[string]bool)
	j := 0
	for _, s := range list {
		if _, ok := seen[s]; !ok {
			list[j] = s
			j++
		}
	}
	result := make([]string, j)
	copy(result, list)
	return result
}

func remove(list []string, element string) []string {
	result := make([]string, 0, len(list))
	for _, s := range list {
		if s != element {
			result = append(result, s)
		}
	}
	return result
}
