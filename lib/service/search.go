package service

import (
	"net/http"
	"github.com/surlykke/RefudeServices/lib/requestutils"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/query"
)

var searchCache = make(map[string][]byte)

// Caller must have mutex
func clearSearchCache() {
	if len(searchCache) > 0 {
		searchCache = make(map[string][]byte)
	}
}

func Search(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	data, ok := searchCache[r.URL.RawQuery]
	if !ok {
		if flatParams, err := requestutils.GetSingleParams(w, r, "type", "q"); err != nil {
			requestutils.ReportUnprocessableEntity(w, err)
			return
		} else {
			mediaType := mediatype.MediaType(flatParams["type"])
			if mediaType == "" {
				mediaType = "application/json"
			}

			var resources = make([]resource.Resource, len(rc))
			var found = 0
			for _, res := range rc {
				if mediatype.MediaTypeMatch(mediaType, res.Mt()) {
					resources[found] = res
					found = found + 1
				}
			}

			resources = resources[:found]
			if q, ok := flatParams["q"]; ok {
				if matcher, err := query.Parse(q); err != nil {
					requestutils.ReportUnprocessableEntity(w, err)
					return
				} else {
					found = 0
					for _, res := range resources {
						if res.Match(matcher) {
							resources[found] = res
							found = found + 1
						}
					}
					resources = resources[:found]
				}
			}

			data = mediatype.ToJSon(resources)
			searchCache[r.URL.RawQuery] = data
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)

}
