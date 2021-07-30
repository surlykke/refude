package doc

import (
	"io/ioutil"
	"net/http"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"

	"github.com/surlykke/RefudeServices/lib/i18n"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type DocType struct {
	Type       string
	Name       string
	NamePlural string
	Collection string
	Doc        string
}

var docTypes = map[string]DocType{
	"application": {
		Type:       "application",
		Name:       i18n.Tr("Application"),
		NamePlural: i18n.Tr("Applications"),
		Collection: "/applications",
		Doc:        "TODO",
	},
	"mimetype": {
		Type:       "mimetype",
		Name:       i18n.Tr("Mimetype"),
		NamePlural: i18n.Tr("Mimetypes"),
		Collection: "/mimetypes",
		Doc:        "TODO",
	},
	"window": {
		Type:       "window",
		Name:       i18n.Tr("Window"),
		NamePlural: i18n.Tr("Windows"),
		Collection: "/windows",
		Doc:        "TODO",
	},
	"file": {
		Type:       "file",
		Name:       i18n.Tr("File"),
		NamePlural: i18n.Tr("Files"),
		Collection: "/file",
		Doc:        "TODO",
	},
	"icontheme": {
		Type:       "icontheme",
		Name:       i18n.Tr("Icontheme"),
		NamePlural: i18n.Tr("Iconthemes"),
		Collection: "/iconthemes",
		Doc:        "TODO",
	},
	"notification": {
		Type:       "notification",
		Name:       i18n.Tr("Notification"),
		NamePlural: i18n.Tr("Notifications"),
		Collection: "/notifications",
		Doc:        "TODO",
	},
	"device": {
		Type:       "device",
		Name:       i18n.Tr("Device"),
		NamePlural: i18n.Tr("Devices"),
		Collection: "/devices",
		Doc:        "TODO",
	},
	"session_action": {
		Type:       "session_action",
		Name:       i18n.Tr("Leave action"),
		NamePlural: i18n.Tr("Leave actions"),
		Collection: "/session/actions",
		Doc:        "TODO",
	},
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/doc" {
		if r.Method == "GET" {
			if bytes, err := ioutil.ReadFile(xdg.DataHome + "/RefudeServices/README.md"); err != nil {
				respond.ServerError(w, err)
			} else {
				w.Header().Set("Content-Type", "text/html")
				w.Write(markdown.ToHTML(bytes, parser.NewWithExtensions(parser.CommonExtensions|parser.AutoHeadingIDs), nil))
			}
		} else {
			respond.NotAllowed(w)
		}
	} else if r.URL.Path == "/doc/types" {
		if r.Method == "GET" {
			respond.AsJson(w, docTypes)
		} else {
			respond.NotAllowed(w)
		}
	} else {
		respond.NotFound(w)
	}
}
