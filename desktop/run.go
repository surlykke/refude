// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package desktop

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/tr"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/search"
	"github.com/surlykke/RefudeServices/statusnotifications"
	"github.com/surlykke/RefudeServices/wayland"
)

//go:embed html
var sources embed.FS

var resourceTemplate *template.Template
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

func init() {
	var bytes []byte
	var err error

	if bytes, err = sources.ReadFile("html/resourceTemplate.html"); err != nil {
		log.Panic(err)
	}
	resourceTemplate = template.Must(template.New("resourceTemplate").Parse(string(bytes)))

	if bytes, err = sources.ReadFile("html/rowTemplate.html"); err != nil {
		log.Panic(err)
	}
	rowTemplate = template.Must(template.New("rowTemplate").Funcs(funcMap).Parse(string(bytes)))

	if bytes, err = sources.ReadFile("html/trayTemplate.html"); err != nil {
		log.Panic(err)
	}
	trayTemplate = template.Must(template.New("trayTemplate").Funcs(funcMap).Parse(string(bytes)))

	if bytes, err = sources.ReadFile("html/menuTemplate.html"); err != nil {
		log.Panic(err)
	}
	menuTemplate = template.Must(template.New("menuTemplate").Funcs(funcMap).Parse(string(bytes)))

}

type item struct {
	IconUrl  string
	ItemPath string
	MenuPath string
}

func init() {
	var tmp http.Handler

	if projectDir, ok := os.LookupEnv("DEV_PROJECT_ROOT_DIR"); ok {
		// Used when developing
		tmp = http.FileServer(http.Dir(projectDir + "/desktop/html"))
	} else if htmlDir, err := fs.Sub(sources, "html"); err == nil {
		// Otherwise, what's baked in
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
			var tabindex = 0

			var term = strings.ToLower(requests.GetSingleQueryParameter(r, "search", ""))
			var expandedResource = requests.GetSingleQueryParameter(r, "details", "")
			var links = search.Search(term)
			var results = make([]result, 0, len(links))
			var resultToFocus = 0
			for i, link := range links {
				tabindex++
				var r = linkAsResult(link, tabindex)
				if r.Path == expandedResource {
					if res := getResource(expandedResource); res != nil {
						var rDet = &resourceDetails{Description: description(res)}
						for _, actionLink := range res.Data().GetLinks(relation.DefaultAction, relation.Action, relation.Delete) {
							tabindex++
							rDet.Actions = append(rDet.Actions, linkAsResult(actionLink, tabindex))
						}
						r.Details = rDet
						resultToFocus = i
					}
				} else {
					r.Comment = link.Type.Short()
				}

				results = append(results, r)
			}
			if len(results) > 0 {
				results[resultToFocus].Autofocus = "autofocus"
			}
			var m = map[string]any{
				"term":    term,
				"results": results,
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
				var iconUrl = i.GetLink(relation.Icon).Href

				var menuPath = resource.GetPath(i.GetLink(relation.Menu))

				items = append(items, item{IconUrl: iconUrl, ItemPath: i.Path, MenuPath: menuPath})
			}
			if err := trayTemplate.Execute(w, map[string]any{"Items": items}); err != nil {
				log.Warn("Error executing bodyTemplate:", err)
			}

		}
	case "/desktop/tray/menu":
		var menuPath = requests.GetSingleQueryParameter(r, "menuPath", "??")
		if menu, ok := repo.Get[*statusnotifications.Menu](menuPath); !ok {
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

func getResource(path string) resource.Resource {
	if strings.HasPrefix(path, "/file/") {
		return file.GetResource(path)
	} else {
		return repo.GetUntyped(path)
	}
}

func linkAsResult(lnk resource.Link, tabindex int) result {
	return result{IconUrl: lnk.IconUrl,
		Title:    lnk.Title,
		Tabindex: tabindex,
		Path:     resource.GetPath(lnk),
		Relation: lnk.Relation}
}

type result struct {
	IconUrl   string
	Title     string
	Tabindex  int
	Path      string
	Relation  relation.Relation
	Autofocus string
	Comment   string
	Details   *resourceDetails
}

type resourceDetails struct {
	Description string
	Actions     []result
}

func description(res resource.Resource) string {
	switch res.Data().Type {
	case mediatype.Window:
		var window = res.(*wayland.WaylandWindow)
		return window.AppId
	case mediatype.Device:
		var dev = res.(*power.Device)
		if dev.Type == "Line Power" {
			if dev.Online {
				return tr.Tr("Plugged in")
			} else {
				return tr.Tr("Not plugged in")
			}
		} else {
			return fmt.Sprintf("%s %d%% - %s", tr.Tr("Level"), dev.Percentage, dev.State)
		}
	default:
		return ""
	}
}

func showBool(b bool) string {
	if b {
		return "yes"
	} else {
		return "no"
	}
}
