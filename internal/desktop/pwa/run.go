// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package pwa

import (
	"bytes"
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/surlykke/refude/internal/lib/resource"
	"github.com/surlykke/refude/internal/lib/respond"
	"github.com/surlykke/refude/internal/lib/utils"
	"github.com/surlykke/refude/internal/search"
)

//go:embed index.html 256.png manifest.json style.css script.js htmx.2.0.10.min.js alpine.3.15.12.min.js
var staticFiles embed.FS

//go:embed searchTemplate.gohtml
var searchTemplateFile embed.FS

//go:embed searchActionsTemplate.gohtml
var searchActionsTemplateFile embed.FS

var searchTemplate *template.Template
var searchActionsTemplate *template.Template
var StaticServer http.Handler

func init() {
	StaticServer = http.StripPrefix("/desktop/", http.FileServer(http.FS(staticFiles)))

	bytes, _ := searchTemplateFile.ReadFile("searchTemplate.gohtml")
	var funcMap = template.FuncMap{
		"safe": func(str string) template.URL { return template.URL(str) },
	}
	searchTemplate = template.Must(template.New("searchTemplate").Funcs(funcMap).Parse(string(bytes)))

	bytes, _ = searchActionsTemplateFile.ReadFile("searchActionsTemplate.gohtml")
	searchActionsTemplate = template.Must(template.New("searchActionsTemplate").Parse(string(bytes)))	
}

func doSearch(w http.ResponseWriter, r *http.Request) {
	var term string = utils.QueryParam(r, "term")
	var rankedList = search.Search(term)
	var b bytes.Buffer
	if err := searchTemplate.Execute(&b, rankedList); err != nil {
		log.Println("Error rendering searchTemplate:", err)
		respond.ServerError(w, err)
	} else {
		respond.AsHtml(w, string(b.Bytes()))
	}
}

func doActions(w http.ResponseWriter, r *http.Request) {
	var path = utils.QueryParam(r, "path")
	var b bytes.Buffer
	if servable, ok := resource.Get[resource.Servable](path); !ok {
		respond.NotFound(w)
	} else  if err := searchActionsTemplate.Execute(&b, servable); err != nil {
		respond.ServerError(w, err)
	} else {
		respond.AsHtml(w, string(b.Bytes()))
	}
}

func Run() {
	http.HandleFunc("GET /desktop/search", doSearch)
	http.HandleFunc("GET /desktop/actions", doActions)
	http.Handle("GET /desktop/", StaticServer)
}
