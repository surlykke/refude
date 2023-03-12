package bookmarks

import (
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/resource"
)

type Bookmarks struct {}

func (bm Bookmarks) Path() string {
	return "/bookmarks"
}

func (bm Bookmarks) Presentation() (string, string, link.Href, string) {
	return "Bookmarks", "", "", "bookmarks"
}

func (bm Bookmarks) Links(context, term string) link.List {
   return link.List{
	{ Href: "/application/", Title: "Applications", Profile: "application*"},
	{ Href: "/window/", Title: "Windows", Profile: "window*"},
	{ Href: "/notification/", Title: "Notifications", Profile: "notification*"},
	{ Href: "/device/", Title: "Devices", Profile: "device*"},
	{ Href: "/item/", Title: "Items", Profile: "item*"}}
}

func (bm Bookmarks) RelevantForSearch() bool {
	return true
}

func Get(path string) resource.Resource {
	return Bookmarks{}
}




