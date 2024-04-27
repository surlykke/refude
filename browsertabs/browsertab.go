package browsertabs

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/resourcerepo"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/stringhash"
	"github.com/surlykke/RefudeServices/watch"
)



type Tab struct {
	resource.BaseResource
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
					var tab = &Tab{BaseResource: *resource.MakeBase("/tab/" +   d["id"], d["title"], d["url"], d["favIcon"], "browsertab")}
					tab.AddLink("", "Focus", tab.IconUrl, relation.Action)
					tabs = append(tabs, tab) 
				}					
				resourcerepo.ReplacePrefixWithList("/tab/", tabs)
				checkForUpdate(tabs) 
				respond.Ok(w)
			}
		}
	} else {
		respond.NotAllowed(w)
	}
}

var Updated atomic.Int64

var hash  atomic.Uint64

func checkForUpdate(tabs []*Tab) {
	var newHash uint64 = 0
	for _, tab := range tabs {
		newHash = newHash ^ stringhash.FNV1a(tab.Base().Title, tab.Base().IconUrl)
	}
	if hash.Swap(newHash) != newHash {
		Updated.Store(time.Now().UnixMicro())
	}
}


