/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package common

import (
	"encoding/json"
	"strings"
	"net/http"
)

type StringList []string

func (pl StringList) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ServeGetAsJson(w, r, pl)
}

func AppendIfNotThere(list StringList, s string) StringList {
	for _, v := range list {
		if v == s {
			return list
		}
	}

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

func Find(list StringList, str string) bool {
	for _,s := range list {
		if s == str {
			return true
		}
	}

	return false
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

func Prepend(stringList StringList, prefix string) StringList {
	res := make(StringList, 0, len(stringList))
	for _,str := range stringList {
		res = append(res, prefix + str)
	}
	return res
}

func ServeGetAsJson(w http.ResponseWriter, r *http.Request, i interface{}) {
	if r.Method == "GET" {
		ServeAsJson(w, r, i)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func ServeAsJson(w http.ResponseWriter, r *http.Request, i interface{}) {
	bytes, err := json.Marshal(i)
	if err != nil {
		panic("Could not json-marshal")
	};
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}


