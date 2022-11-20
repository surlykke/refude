package bookmarks

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
)

var bookmarks = resource.Wrapper {
	Self: "/bookmarks/",
	Links: link.List{
	{ Href: "/application/", Title: "Applications", Profile: "application*"},
	{ Href: "/window/", Title: "Windows", Profile: "window*"},
	{ Href: "/notification/", Title: "Notifications", Profile: "notification*"},
	{ Href: "/device/", Title: "Devices", Profile: "device*"},
	{ Href: "/item/", Title: "Items", Profile: "item*"}},
	Title: "Bookmarks",
	Profile: "bookmarks",
}


func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, bookmarks)
	} else {
		respond.NotAllowed(w)
	}
}

