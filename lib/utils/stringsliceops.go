/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package utils

import (
	"strings"
)


func Contains(sl []string, s string) bool {
	for _, str := range sl {
		if str == s {
			return true
		}
	}

	return false
}

func AppendIfNotThere(list []string, s string) []string {
	for _, v := range list {
		if v == s {
			return list
		}
	}

	return append(list, s)
}

func PushFront(s string, list []string) []string {
	res := make([]string, 1 + len(list))
	res[0] = s;
	for i,item := range list {
		res[i + 1] = item
	}

	return res
}

func Remove(list []string, str string) []string {
	res := make([]string, 0, len(list))
	for _,s := range list {
		if s != str {
			res = append(res, s)
		}
	}
	return res
}


func Split(str string, sep string) []string {
	tmp := strings.Split(str, sep)
	res := make([]string, 0, len(tmp))
	for _,s := range tmp {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			res = AppendIfNotThere(res, trimmed)
		}
	}
	return res
}

func PrependEach(sl []string, prefix string) []string {
	res := make([]string, 0, len(sl))
	for _,str := range sl {
		res = append(res, prefix + str)
	}
	return res
}



