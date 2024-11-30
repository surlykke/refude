// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package desktop

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/lib/icon"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/path"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/search"
	"github.com/surlykke/RefudeServices/statusnotifications"
)

//go:embed html
var sources embed.FS

var rowTemplate *template.Template
var trayTemplate *template.Template
var menuTemplate *template.Template
var StaticServer http.Handler

var funcMap = template.FuncMap{
	// The name "inc" is what the function will be called in the template text.
	"inc": func(i int) int {
		return i + 1
	},
}

func loadTemplate(name, relPath string) *template.Template {
	var bytes []byte
	var err error
	if bytes, err = sources.ReadFile(relPath); err != nil {
		log.Panic(err)
	}
	return template.Must(template.New(name).Funcs(funcMap).Parse(string(bytes)))
}

func init() {
	rowTemplate = loadTemplate("rowTemplate", "html/rowTemplate.html")
	trayTemplate = loadTemplate("trayTemplate", "html/trayTemplate.html")
	menuTemplate = loadTemplate("menuTemplate", "html/menuTemplate.html")
}

type item struct {
	Icon     icon.Name
	ItemPath path.Path
	MenuPath path.Path
}

func init() {
	var tmp http.Handler

	if htmlDir, err := fs.Sub(sources, "html"); err == nil {
		tmp = http.FileServer(http.FS(htmlDir))
	} else {
		log.Panic(err)
	}
	StaticServer = http.StripPrefix("/desktop", tmp)

}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/desktop/search":
		if r.Method != "GET" {
			respond.NotAllowed(w)
		} else {
			var term = requests.GetSingleQueryParameter(r, "search", "")
			var m = map[string]any{
				"term":  term,
				"links": search.Search(term),
			}
			if err := rowTemplate.Execute(w, m); err != nil {
				log.Warn("Error executing rowTemplate:", err)
			}
		}
	case "/desktop/tray":
		if r.Method != "GET" {
			respond.NotAllowed(w)
		} else {
			var items = make([]item, 0, 10)
			for _, i := range repo.GetListSortedByPath[*statusnotifications.Item]("/item/") {
				items = append(items, item{Icon: i.Link().Icon, ItemPath: i.Path, MenuPath: i.MenuPath})
			}
			if err := trayTemplate.Execute(w, map[string]any{"Items": items}); err != nil {
				log.Warn("Error executing bodyTemplate:", err)
			}

		}
	case "/desktop/tray/menu":
		var menuPath = requests.GetSingleQueryParameter(r, "menuPath", "??")
		if menu, ok := repo.Get[*statusnotifications.Menu](path.Of(menuPath)); !ok {
			respond.NotFound(w)
		} else if entries, err := menu.Entries(); err != nil {
			respond.ServerError(w, err)
		} else {
			if err := menuTemplate.Execute(w, entries); err != nil {
				log.Warn("Error executing menuTemplate:", err)
			}
		}
	default:
		if strings.HasSuffix(r.URL.Path, "Template.html") {
			respond.NotFound(w)
		} else {
			StaticServer.ServeHTTP(w, r)
		}
	}
}
