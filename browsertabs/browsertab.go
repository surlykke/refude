package browsertabs

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/watch"
)

type Tab struct {
	resource.BaseResource
}

func (this *Tab) Id() string {
	return this.Path[len("/tab/"):]
}

func (this *Tab) DoPost(w http.ResponseWriter, r *http.Request) {
	watch.Publish("focusTab", this.Id())
	respond.Accepted(w)
}

func (this *Tab) RelevantForSearch() bool {
	return !strings.HasPrefix(this.Title, "Refude launcher")
}


var Tabs = resource.MakeCollection[*Tab]()


func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/tab/" && r.Method == "POST" {
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
					tabs = append(tabs, &Tab{
						BaseResource: resource.BaseResource{
							Path:    "/tab/" +   d["id"],
							Title:   d["title"],
							Comment: d["url"],
							IconUrl: link.Href(d["favIcon"]),
							Profile: "browsertab",
						}})
				}
				Tabs.ReplaceWith(tabs)
				respond.Ok(w)
			}
		}
	} else {
		Tabs.ServeHTTP(w, r)
	}
}
