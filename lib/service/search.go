package service

import (
	"github.com/surlykke/RefudeServices/lib/resource"
	"net/http"
	"github.com/surlykke/RefudeServices/lib/query"
	"github.com/surlykke/RefudeServices/lib/requestutils"
	"github.com/surlykke/RefudeServices/lib/mediatype"
)

func Search(w http.ResponseWriter, r *http.Request) {
	if flatParams, ok := requestutils.GetSingleParams(w, r, "type", "q"); ok {
		mediaType := mediatype.MediaType(flatParams["type"])
		if mediaType == "" {
			mediaType = "application/json"
		}
		var result = collectByType(mediaType)
		if q, ok := flatParams["q"]; ok {
			if matcher, ok2 := requestutils.GetMatcher(w, q); ok2 {
				result = filter(result, matcher)
			} else {
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(mediatype.ToJSon(result))
	}

}

func collectByType(mediaType mediatype.MediaType) []resource.Resource {
	mutex.Lock()
	defer mutex.Unlock()

	var result = make([]resource.Resource, len(rc))
	var found = 0
	for _,res := range rc {
		if mediatype.MediaTypeMatch(mediaType, res.Mt()) {
			result[found] = res
			found = found + 1
		}
	}

	return result[:found]
}

// Messes up it's argument, don't use it afterwards
func filter(resources []resource.Resource, matcher query.Matcher) []resource.Resource {
	var pos = 0
	for _, res := range resources {
		if res.Match(matcher) {
			resources[pos] = res
			pos = pos + 1
		}
	}

	return resources[:pos]
}

