// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package requests

import (
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/lib/parser"
)

func GetSingleQueryParameter(r *http.Request, parameterName string, fallbackValue string) string {
	if len(r.URL.Query()[parameterName]) == 0 {
		return fallbackValue
	} else {
		return r.URL.Query()[parameterName][0]
	}
}

func Term(r *http.Request) string {
	return strings.ToLower(strings.TrimSpace(GetSingleQueryParameter(r, "term", "")))
}

func GetMatcher(r *http.Request) (parser.Matcher, error) {
	if q, ok := r.URL.Query()["q"]; ok && len(q) > 0 {
		return parser.Parse(q[0])
	} else {
		return nil, nil
	}
}

func HaveParam(r *http.Request, paramName string) bool {
	var _, ok = r.URL.Query()[paramName]
	return ok
}
