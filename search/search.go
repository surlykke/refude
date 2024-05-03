package search

import (
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/resourcerepo"
	"github.com/surlykke/RefudeServices/lib/respond"
)

func Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respond.NotAllowed(w)
	} else {
		var from = requests.GetSingleQueryParameter(r, "from", "/start")
		if res := FetchResource(from); res == nil {
			respond.NotFound(w)
		} else {
			var searchTerm = requests.GetSingleQueryParameter(r, "search", "")
			if searchable, ok := res.(resource.Searchable); ok {
				resource.ServeList(w, r, searchable.Search(searchTerm))
			} else {
				resource.ServeList(w, r, []resource.Resource{})
			}
		}
	}
}


func FetchResource(path string) resource.Resource {
	if strings.HasPrefix(path, "/file/") {
		return file.FileRepo.GetResource(path[5:])
	} else if res, ok := resourcerepo.Get(path); ok {
		return res
	} else {
		return nil
	}
}
