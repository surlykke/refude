// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package requests

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/surlykke/RefudeServices/lib/parser"
	"net/http"
	"regexp"
)

func GetSingleQueryParameter(r *http.Request, parameterName string, fallbackValue string) string {
	if len(r.URL.Query()[parameterName]) == 0 {
		return fallbackValue
	} else {
		return r.URL.Query()[parameterName][0]
	}
}

/**
 * Errors if r.URL.Query contains parameters not in params
 * Errors if any of paramNames has more than one value
 * Returns value from query if there, "" if not
 */
func GetSingleParams(r *http.Request, paramNames ...string) (map[string]string, error) {
	for queryParam, _ := range r.URL.Query() {
		var ok = false
		for _, paramName := range paramNames {
			if queryParam == paramName {
				ok = true
				break
			}
		}

		if !ok {
			return nil, errors.Errorf("Unexpected parameter: %s", queryParam)
		}
	}
	var result = make(map[string]string)

	for _, paramName := range paramNames {
		if len(r.URL.Query()[paramName]) > 1 {
			return nil, errors.Errorf("Multiple parameter value for %s", paramNames)
		} else if len(r.URL.Query()[paramName]) == 1 {
			result[paramName] = r.URL.Query()[paramName][0]
		}

	}

	return result, nil
}


func ReportUnprocessableEntity(w http.ResponseWriter, err error) {
	fmt.Println("unp: err: ", err)
	if body, err2 := json.Marshal(err.Error()); err2 == nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(body)
	} else {
		panic(fmt.Sprintf("Cannot json-marshall %s", err.Error()))
	}
}

var r = regexp.MustCompile(`^\s*(?:W/)?("[^"]*")\s*`)

// We do not do weak matches, so any 'W/' preceding a tag is
// ignored
func EtagMatch(etag string, etagList string) bool {
	var pos = 0

	for {
		matched := r.FindStringSubmatch(etagList[pos:])
		if matched != nil {
			if etag == matched[1] {
				return true
			} else {
				pos += len(matched[0])
			}
		} else {
			return false
		}

		if pos >= len(etagList) {
			return false
		} else {
			if etagList[pos] != ',' {
				return false
			} else {
				pos++
			}
		}
	}
}

func MatchAny(interface{}) bool {
	return true
}

func GetMatcher(r *http.Request) (parser.Matcher, error) {
	if q, ok := r.URL.Query()["q"]; ok && len(q) > 0 {
		fmt.Println("query:", q, "returning parsed matcher")
		return parser.Parse(q[0])
	} else {
		fmt.Println("No query, returning matchAny")
		return MatchAny, nil
	}
}

func GetMatcher2(r *http.Request) (parser.Matcher, error) {
	if q, ok := r.URL.Query()["q"]; ok && len(q) > 0 {
		return parser.Parse(q[0])
	} else {
		return nil, nil
	}
}
