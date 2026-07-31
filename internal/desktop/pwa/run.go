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

	"github.com/surlykke/refude/internal/lib/respond"
	"github.com/surlykke/refude/internal/lib/utils"
	"github.com/surlykke/refude/internal/search"
)

//go:embed index.html 256.png manifest.json style.css script.js htmx.2.0.10.min.js alpine.3.15.12.min.js
 
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

func doSearch(w http.ResponseWriter, r *http.Request) {
	var term string = utils.QueryParam(r, "term")
	var resources = search.Search(term)
	var b bytes.Buffer
	if err := searchTemplate.Execute(&b, resources); err != nil {
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
