package service

import (
	"github.com/surlykke/RefudeServices/lib/resource"
	"net/http"
	"fmt"
	"github.com/surlykke/RefudeServices/lib/query"
)

type Search struct{
	resource.DefaultResource
}

func (s *Search) GET(w http.ResponseWriter, r *http.Request) {
	var flatParams, err = resource.GetSingleParams(r, "type", "q")
	if err != nil {
		fmt.Println("Error in search.GET: ", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	var mediaType = resource.MediaType(flatParams["type"])
	var matcher query.Matcher
	if q := flatParams["q"]; q == "" {
		matcher = func(res interface{}) bool { return true }
	} else {
		if matcher, err = query.Parse(q); err != nil {
			fmt.Println("Query error: ", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resource.ToJSon(collectResources(mediaType, matcher)))
}

func collectResources(mediaType resource.MediaType, matcher query.Matcher) []resource.Resource {
	mutex.Lock()
	defer mutex.Unlock()

	var result = make([]resource.Resource, 0, len(rc))
	for _,res := range rc {
		if (mediaType == "" || mediaType == res.MediaType()) && matcher(res) {
			result = append(result, res)
		}
	}

	return result
}

func (s *Search) MediaType() resource.MediaType {
	return resource.MediaType("application/json")
}
