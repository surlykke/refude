package browser

import (
	"net/url"
	"strings"

	"github.com/surlykke/RefudeServices/server/applications"
	"github.com/surlykke/RefudeServices/server/lib/entity"
	"github.com/surlykke/RefudeServices/server/lib/response"
)

type Tab struct {
	entity.Base
	Id        string
	BrowserId string
	Url       string
}

func (this *Tab) DoPost(action string) response.Response {
	if app, ok := applications.AppMap.Get(this.BrowserId); ok {
		app.Run("http://refude.focustab.localhost?url=" + url.QueryEscape(this.Url))
		return response.Accepted()
	} else {
		return response.NotFound()
	}

}

func (this *Tab) OmitFromSearch() bool {
	return strings.HasPrefix(this.Url, "http://localhost:7938/desktop")
}
