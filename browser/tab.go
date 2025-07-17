package browser

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/surlykke/RefudeServices/lib/entity"
	"github.com/surlykke/RefudeServices/lib/icon"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/response"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/surlykke/RefudeServices/watch"
)

type sinkData struct {
	Id      string    `json:"id"`
	Url     string    `json:"url"`
	Title   string    `json:"title"`
	Favicon icon.Name `json:"favicon"`
}

func TabsDoPost(browserName string, dataList []sinkData) response.Response {
	fmt.Println("TabsDoPost")
	var mapOfTabs = make(map[string]*Tab, len(dataList))
	for _, d := range dataList {
		if len(d.Title) > 60 { // Shorten title a bit
			d.Title = d.Title[0:60] + "..."
		}
		var tab = &Tab{Base: *entity.MakeBase(d.Title, "Chrome tab", d.Favicon, mediatype.Tab), Id: d.Id, BrowserId: browserName, Url: d.Url}
		tab.AddAction("", browserName+" tab", "")
		//tab.AddDeleteAction("close", title, "Close tab", "")

		mapOfTabs[d.Id] = tab
	}

	TabMap.Replace(mapOfTabs, func(t *Tab) bool { return t.BrowserId == browserName })
	watch.Publish("search", "")
	return response.Accepted()
}

type Tab struct {
	entity.Base
	Id        string
	BrowserId string
	Url       string
}

func (this *Tab) DoPost(action string) response.Response {
	xdg.RunCmd("google-chrome", fmt.Sprintf("http://refudecommand.localhost?url=%s&action=focus", url.QueryEscape(this.Url)))
	return response.Accepted()
}

func (this *Tab) DoDelete() response.Response {
	commands <- tabCommand{BrowserName: this.BrowserId, Operation: "delete", TabId: this.Id}
	return response.Accepted()
}

func (this *Tab) OmitFromSearch() bool {
	return strings.HasPrefix(this.Url, "http://localhost:7938/desktop")
}
