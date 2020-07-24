package doc

import (
	"io/ioutil"
	"net/http"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	. "github.com/surlykke/RefudeServices/lib/i18n"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

func Handler(r *http.Request) http.Handler {
	if r.URL.Path == "/doc" {
		return readme
	} else if r.URL.Path == "/doc/types" {
		return DocTypes
	} else {
		return nil
	}
}

type Readme struct{}

func (d Readme) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
}

var readme Readme

type DocType struct {
	Type       string
	Name       string
	NamePlural string
	Collection string
	Doc        string
}

type TypeMap map[string]DocType

func (tm TypeMap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, &respond.StandardFormat{
			Self:  "/doc/types",
			Type:  "doctypes",
			Title: "Refude document types",
			Data:  tm,
		})
	} else {
		respond.NotAllowed(w)
	}

}

var DocTypes = TypeMap{
	"application": {
		Type:       "application",
		Name:       Tr("Application"),
		NamePlural: Tr("Applications"),
		Collection: "/applications",
		Doc:        "TODO",
	},
	"mimetype": {
		Type:       "mimetype",
		Name:       Tr("Mimetype"),
		NamePlural: Tr("Mimetypes"),
		Collection: "/mimetypes",
		Doc:        "TODO",
	},
	"window": {
		Type:       "window",
		Name:       Tr("Window"),
		NamePlural: Tr("Windows"),
		Collection: "/windows",
		Doc:        "TODO",
	},
	"file": {
		Type:       "file",
		Name:       Tr("File"),
		NamePlural: Tr("Files"),
		Collection: "/file",
		Doc:        "TODO",
	},
	"icontheme": {
		Type:       "icontheme",
		Name:       Tr("Icontheme"),
		NamePlural: Tr("Iconthemes"),
		Collection: "/iconthemes",
		Doc:        "TODO",
	},
	"notification": {
		Type:       "notification",
		Name:       Tr("Notification"),
		NamePlural: Tr("Notifications"),
		Collection: "/notifications",
		Doc:        "TODO",
	},
	"device": {
		Type:       "device",
		Name:       Tr("Device"),
		NamePlural: Tr("Devices"),
		Collection: "/devices",
		Doc:        "TODO",
	},
	"session_action": {
		Type:       "session_action",
		Name:       Tr("Leave action"),
		NamePlural: Tr("Leave actions"),
		Collection: "/session/actions",
		Doc:        "TODO",
	},
}
