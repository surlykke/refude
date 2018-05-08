package requestutils

import (
	"regexp"
	"net/http"
	"github.com/surlykke/RefudeServices/lib/query"
	"fmt"
	"github.com/surlykke/RefudeServices/lib/mediatype"
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

func GetMatcher(w http.ResponseWriter, q string) (query.Matcher, bool) {
	if matcher, err := query.Parse(q); err != nil {
		ReportUnprocessableEntity(w, "Error in query: %s", mediatype.ToJSon(err))
		return nil, false
	} else {
		return matcher, true
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
func GetSingleParams(w http.ResponseWriter, r *http.Request, paramNames ...string) (map[string]string, bool) {
	for queryParam, _ := range r.URL.Query() {
		var ok = false
		for _, paramName := range paramNames {
			if queryParam == paramName {
				ok = true
				break
			}
		}

		if !ok {
			ReportUnprocessableEntity(w, "Unexpected parameter: %s", queryParam)
			return nil, false
		}
	}
	var result = make(map[string]string)

	for _, paramName := range paramNames {
		if len(r.URL.Query()[paramName]) > 1 {
			ReportUnprocessableEntity(w, "Multiple parameter value for %s", paramNames)
			return nil, false
		} else if len(r.URL.Query()[paramName]) == 1 {
			result[paramName] = r.URL.Query()[paramName][0]
		}

	}

	return result, true
}

func ReportUnprocessableEntity(w http.ResponseWriter, errMsgFmt string, args...interface{}) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	w.Write(mediatype.ToJSon(fmt.Sprintf(errMsgFmt, args...)))
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



