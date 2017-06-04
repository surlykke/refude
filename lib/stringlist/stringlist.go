/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package stringlist

import (
	"strings"
	"net/http"
	"github.com/surlykke/RefudeServices/lib/common"
)

type StringList []string

func (sl StringList) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ServeGetAsJson(w, r, sl)
}

func (sl StringList) Has(s string) bool {
	for _, str := range sl {
		if str == s {
			return true
		}
	}

	return false
}

func AppendIfNotThere(list StringList, s string) StringList {
	for _, v := range list {
		if v == s {
			return list
		}
	}

	return append(list, s)
}

func PushFront(s string, list StringList) StringList {
	res := make(StringList, 1 + len(list))
	res[0] = s;
	for i,item := range list {
		res[i + 1] = item
	}

	return res
}

func PushBack(list StringList, s string) StringList {
	return append(list, s)
}

func Remove(list StringList, str string) StringList {
	res := make(StringList, 0, len(list))
	for _,s := range list {
		if s != str {
			res = append(res, s)
		}
	}
	return res
}


func Split(str string, sep string) StringList {
	tmp := strings.Split(str, sep)
	res := make(StringList, 0, len(tmp))
	for _,s := range tmp {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			res = AppendIfNotThere(res, trimmed)
		}
	}
	return res
}

func PrependEach(stringList StringList, prefix string) StringList {
	res := make(StringList, 0, len(stringList))
	for _,str := range stringList {
		res = append(res, prefix + str)
	}
	return res
}

func ServeGetAsJson(w http.ResponseWriter, r *http.Request, i interface{}) {
	if r.Method == "GET" {
		common.ServeAsJson(w, r, i)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}



