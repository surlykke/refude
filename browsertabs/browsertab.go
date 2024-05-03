package browsertabs

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/resourcerepo"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/watch"
)



type Tab struct {
	resource.ResourceData
}

func (this *Tab) Id() string {
	return this.Path[len("/tab/"):]
}

func (this *Tab) RelevantForSearch(searchTerm string) bool {
	return this.Title != "Refude Desktop" && !strings.HasPrefix(this.Comment, "http://localhost:7938/desktop")
}

func (this *Tab) DoPost(w http.ResponseWriter, r *http.Request) {
	watch.Publish("focusTab", this.Id())
	respond.Accepted(w)
}

func Search(list *resource.RRList, term string) {
	for _, tab := range resourcerepo.GetTypedByPrefix[*Tab]("/tab/") {
		if ! strings.HasPrefix(tab.Comment, "http://localhost:7938/desktop") {
			if rnk := searchutils.Match(term, tab.Title); rnk >= 0 {
				*list = append(*list, resource.RankedResource{Res: tab, Rank: rnk})
			}
		}
	}
}


func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if r.Body == nil {
			respond.UnprocessableEntity(w, errors.New("No data"))
		} else if bytes, err := io.ReadAll(r.Body); err != nil {
			respond.ServerError(w, err)
		} else {
			var data = make([]map[string]string, 30)
			if err := json.Unmarshal(bytes, &data); err != nil {
				respond.UnprocessableEntity(w, err)
			} else {
				var tabs = make([]*Tab, 0, len(data))
				for _, d := range data {
					var title = d["title"] 
					if len(title) > 60 { // Shorten title a bit
						if spacePos := strings.Index(title[60:], " "); spacePos > -1 {
							title = title[0: 60 + spacePos] + "..."
						} else {
							title = title[0:60] + "..."
						}
					}
					var url = d["url"]
					if queryStart := strings.Index(url, "?"); queryStart > -1 {
						url = url[0:queryStart + 1] + "..."
					} else if len(url) > 60 {
						url = url[0:60] + "..."
					}

					var tab = &Tab{ResourceData: *resource.MakeBase("/tab/" +   d["id"], title, url, d["favIcon"], "tab")}
					tab.AddLink("", "Focus", tab.IconUrl, relation.Action)
					tabs = append(tabs, tab) 
				}					
				resourcerepo.ReplacePrefixWithList("/tab/", tabs)
				respond.Ok(w)
			}
		}
	} else {
		respond.NotAllowed(w)
	}
}

