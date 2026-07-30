// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package pwa

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/internal/lib/respond"
	"github.com/surlykke/refude/internal/lib/utils"
	"github.com/surlykke/refude/internal/search"
)

//go:embed index.html 256.png htmx.min.js manifest.json script.js style.css
var staticFiles embed.FS

//go:embed searchTemplate.html
var searchTemplateFile embed.FS

var searchTemplate *template.Template

var StaticServer http.Handler

func init() {
	StaticServer = http.StripPrefix("/desktop/", http.FileServer(http.FS(staticFiles)))

	bytes, _ := searchTemplateFile.ReadFile("searchTemplate.html")
	fmt.Println("searchTemplate.html", len(bytes), "bytes")
	searchTemplate = template.Must(template.New("searchTemplate").Parse(string(bytes)))
}

type Resourceline struct {
	Icon        string
	Title       string
	SubTitle    string
	Description string
	Href        string
	AltActions  []AltAction
}

type AltAction struct {
	Name string
	Href string
}

func doSearch(w http.ResponseWriter, r *http.Request) {
	var lines = make([]Resourceline, 0, 50)
	var term string = utils.QueryParam(r, "term")
	for _, r := range search.Search(term) {
		var line = Resourceline{Icon: string(r.Icon), Title: r.Title, SubTitle: r.Subtitle}
		var links = r.GetLinks(entity.OrgRefudeAction)
		if len(links) > 0 {
			line.Href = links[0].Href
			if len(links) > 1 {
				for _, l := range links[1:] {
					line.AltActions = append(line.AltActions, AltAction{Name: l.Title, Href: l.Href})
				}
			}
		}
		if "Notification" == r.Kind {
			line.Description = r.Comment
		}
		lines = append(lines, line)
	}

	var b bytes.Buffer
	if err := searchTemplate.Execute(&b, lines); err != nil {
		log.Println("Error rendering searchTemplate:", err)
		respond.ServerError(w, err)
	} else {
		respond.AsHtml(w, string(b.Bytes()))
	}
}

func Run() {
	http.HandleFunc("GET /desktop/search", doSearch)
	http.Handle("GET /desktop/", StaticServer)
}
