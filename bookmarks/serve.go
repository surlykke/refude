package bookmarks

import (
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/resource"
)

type Bookmarks struct {
	resource.BaseResource
}


func (bm Bookmarks) Links(searchTerm string) link.List {
   return link.List{
	{ Href: "/application/", Title: "Applications", Profile: "application*"},
	{ Href: "/window/", Title: "Windows", Profile: "window*"},
	{ Href: "/notification/", Title: "Notifications", Profile: "notification*"},
	{ Href: "/device/", Title: "Devices", Profile: "device*"},
	{ Href: "/item/", Title: "Items", Profile: "item*"}}
}

func Get(path string) resource.Resource {
	return &Bookmarks{BaseResource:resource.BaseResource{Path: "/bookmarks", Title: "Bookmarks", Profile: "bookmarks"}}
}




