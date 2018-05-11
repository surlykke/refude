// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package requestutils

import (
	"regexp"
	"net/http"
	"fmt"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/pkg/errors"
)

var r = regexp.MustCompile(`^\s*(?:W/)?("[^"]*")\s*`)


func extractETags(s string) []string {
	var result = make([]string, 0, 5)
	var pos = 0

	for {
		matched := r.FindStringSubmatch(s[pos:])
		if matched != nil {
			result = append(result, matched[1])
			pos += len(matched[0])
		} else {
			return nil
		}

		if pos >= len(s) {
			return result
		} else {
			if s[pos] != ',' {
				return nil
			} else {
				pos++
			}
		}
	}
}


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
func GetSingleParams(w http.ResponseWriter, r *http.Request, paramNames ...string) (map[string]string, error) {
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
	w.WriteHeader(http.StatusUnprocessableEntity)
	w.Write(mediatype.ToJSon(err.Error()))
}



/*
 * Will return false if
 *  - request contains no valid If-None-Match or
 *  - none of the etags in the If-None-Match equals the given etag
 *
 * We do not differ between hard and weak etags - iow. we ignore W/ prefixes
 */
func EtagMatch(r *http.Request, etag string) bool {
	fmt.Println("EtagMatch")
	if ifNoneMatch := r.Header.Get("if-none-match"); ifNoneMatch != "" {
		fmt.Println("ifNoneMatch: ", ifNoneMatch)
		if tags := extractETags(ifNoneMatch); tags != nil {
			for _, tag := range tags {
				fmt.Println("Compare", tag, "to", etag)
				if etag == tag {
					return true
				}
			}
		}
	}

	return false
}



