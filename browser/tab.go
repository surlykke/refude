package browser

import (
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/watch"
)

type Tab struct {
	resource.ResourceData
	Url string
}

func (this *Tab) Id() string {
	return string(this.Path[len("/tab/"):])
}

func (this *Tab) DoPost(w http.ResponseWriter, r *http.Request) {
	watch.Publish("focusTab", this.Id())
	respond.Accepted(w)
}

func (this *Tab) DoDelete(w http.ResponseWriter, r *http.Request) {
	watch.Publish("closeTab", this.Id())
	respond.Accepted(w)
}

func (this *Tab) OmitFromSearch() bool {
	return strings.HasPrefix(this.Url, "http://localhost:7938/desktop")
}
