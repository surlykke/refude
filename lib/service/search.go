// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
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
		if flatParams, err := requestutils.GetSingleParams(r, "type", "q"); err != nil {
			requestutils.ReportUnprocessableEntity(w, resource.ToJSon(err))
			return
		} else {
			mediaType := mediatype.MediaType(flatParams["type"])
			if mediaType == "" {
				mediaType = "application/json"
			}

			var jsonResources = make([]*resource.JsonResource, len(rc))
			var found = 0
			for _, jsonRes := range rc {
				if mediatype.MediaTypeMatch(mediaType, jsonRes.GetMt()) {
					jsonResources[found] = jsonRes
					found = found + 1
				}
			}

			jsonResources = jsonResources[:found]
			if q, ok := flatParams["q"]; ok {
				if matcher, err := query.Parse(q); err != nil {
					requestutils.ReportUnprocessableEntity(w, resource.ToJSon(err))
					return
				} else {
					found = 0
					for _, jsonRes := range jsonResources {
						if jsonRes.Matches(matcher) {
							jsonResources[found] = jsonRes
							found = found + 1
						}
					}
					jsonResources = jsonResources[:found]
				}
			}

			for _, jsonRes := range jsonResources {
				jsonRes.EnsureReady()
			}
			data = resource.ToJSon(jsonResources)
			searchCache[r.URL.RawQuery] = data
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)

}
