package browser

import (
	"strings"

	"github.com/surlykke/RefudeServices/lib/entity"
	"github.com/surlykke/RefudeServices/lib/response"
	"github.com/surlykke/RefudeServices/watch"
)

type Tab struct {
	entity.Base
	Id  string
	Url string
}

func (this *Tab) DoPost(action string) response.Response {
	watch.Publish("focusTab", this.Id)
	return response.Accepted()
}

func (this *Tab) DoDelete() response.Response {
	watch.Publish("closeTab", this.Id)
	return response.Accepted()
}

func (this *Tab) OmitFromSearch() bool {
	return strings.HasPrefix(this.Url, "http://localhost:7938/desktop")
}
