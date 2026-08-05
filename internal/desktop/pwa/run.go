// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package pwa

import (
	"bytes"
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/internal/lib/respond"
	"github.com/surlykke/refude/internal/lib/utils"
	"github.com/surlykke/refude/internal/search"
)

//go:embed index.html 256.png manifest.json style.css script.js htmx.2.0.10.min.js alpine.3.15.12.min.js
var staticFiles embed.FS

//go:embed searchTemplate.gohtml
var searchTemplateFile embed.FS

var searchTemplate *template.Template
var StaticServer http.Handler

func init() {
	StaticServer = http.StripPrefix("/desktop/", http.FileServer(http.FS(staticFiles)))

	bytes, _ := searchTemplateFile.ReadFile("searchTemplate.gohtml")
	searchTemplate = template.Must(template.New("searchTemplate").Parse(string(bytes)))
}

type View struct {
	Title        string
	Summary      string
	Icon         string
	PostAction   string
	OtherActions []Action
}

type Action struct {
	Title      string
	PostAction string
}

func Make(base entity.Base) View {
	var v = View{
		Title:   base.Title,
		Summary: base.Summary,
		Icon:    base.Icon,
	}
	if len(base.Actions) > 0 {
		v.PostAction = base.Path + "?action=" + base.Actions[0].Id
	}
	for i, a := range base.Actions {
		if i > 0 {
			v.OtherActions = append(v.OtherActions, Action{Title: a.Name, PostAction: base.Path + "?action=" + a.Id})
		}
	}
	return v
}

func doSearch(w http.ResponseWriter, r *http.Request) {
	var term string = utils.QueryParam(r, "term")
	var rankedList = search.Search(term)
	var views = make([]View, 0, len(rankedList))
	for _, ranked := range rankedList {
		views = append(views, Make(ranked.Base))
	}
	var b bytes.Buffer
	if err := searchTemplate.Execute(&b, views); err != nil {
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
