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
	"os"

	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/search"
	"github.com/surlykke/RefudeServices/watch"
	"github.com/surlykke/RefudeServices/wayland"
	"golang.org/x/exp/slices"
)

// Constants
var headingOrder = map[string]int{"Actions": 0, "Notifications": 1, "Windows": 2, "Tabs": 3, "Applications": 4, "Files": 5, "Other": 6}

var profileHeadingMap = map[string]string{
	"notification": "Notifications", "window": "Windows", "browsertab": "Tabs", "application": "Applications", "file": "Files",
}

//

//go:embed html
var sources embed.FS

var mainTemplate *template.Template
var StaticServer http.Handler

func init() {
	var bytes []byte
	var err error

	if bytes, err = sources.ReadFile("html/mainTemplate.html"); err != nil {
		log.Panic(err)
	}

	mainTemplate = template.Must(template.New("mainTemplate").Funcs(template.FuncMap{"trClass": trClass}).Parse(string(bytes)))
}

type trRow struct {
	Heading  string
	Class    string
	IconUrl  string
	Title    string
	Href     string
	Relation relation.Relation
	Profile  string
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
	case "/desktop/", "/desktop/index.html":
		var resourcePath = requests.GetSingleQueryParameter(r, "resource", "/start")

		if res := search.FetchResource(resourcePath); res != nil {
			var (
				term         = requests.GetSingleQueryParameter(r, "search", "")
				selected, _  = requests.GetPosInt(r, "selected")
				actions      = res.Base().ActionLinks(term)
				subresources = res.Search(term)
				trRows       = make([]trRow, 0, len(actions)+len(subresources))
			)

			for _, a := range actions {
				trRows = append(trRows, trRow{Heading: "Action", IconUrl: a.IconUrl, Title: a.Title, Href: a.Href, Relation: a.Relation, Class: "selectable"})
			}
			for _, sr := range subresources {
				var row = trRow{IconUrl: sr.Base().IconUrl, Title: sr.Base().Title, Href: sr.Base().Path, Relation: relation.Self, Profile: sr.Base().Profile, Class: "selectable"}
				if heading, ok := profileHeadingMap[sr.Base().Profile]; ok {
					row.Heading = heading
				} else {
					row.Heading = "Other"
				}
				trRows = append(trRows, row)
			}
			slices.SortFunc(trRows, func(r1, r2 trRow) bool { return headingOrder[r1.Heading] < headingOrder[r2.Heading] })

			if len(trRows) > 0 {
				var lastHeading string

				for i := 0; i < len(trRows); i++ {
					if trRows[i].Heading != lastHeading {
						lastHeading = trRows[i].Heading
					} else {
						trRows[i].Heading = ""
					}
				}
				if int(selected) >= len(trRows) {
					selected = 0
				}
				trRows[selected].Class = trRows[selected].Class + " selected"
			}
			var m = map[string]any{
				"Searchable": res.Base().Searchable(),
				"Title":      res.Base().Title,
				"Icon":       res.Base().IconUrl,
				"Term":       term,
				"Rows":       trRows,
			}

			if err := mainTemplate.Execute(w, m); err != nil {
				log.Warn("Error executing mainTemplate:", err)
			}

		} else {
			respond.NotFound(w)
		}

	case "/desktop/show":
		if r.Method != "POST" {
			respond.NotAllowed(w)
		} else {
			wayland.RememberActive()
			watch.Publish("showDesktop", "")
			respond.Accepted(w)
		}
	case "/desktop/hide":
		if r.Method != "POST" {
			respond.NotAllowed(w)
		} else {
			switch requests.GetSingleQueryParameter(r, "restore", "") {
			case "window":
				wayland.ActivateRememberedActive()
				fallthrough
			case "tab":
				watch.Publish("restoreTab", "")
			}
			watch.Publish("hideDesktop", "")
			respond.Accepted(w)
		}
	case "mainTemplate.html":
		respond.NotFound(w)
	default:
		StaticServer.ServeHTTP(w, r)
	}
}

func trClass(selected string, pathOrHref string) string {
	if selected == pathOrHref {
		return "selectable selected"
	} else {
		return "selectable"
	}
}
