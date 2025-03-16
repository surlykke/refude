// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package desktop

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/icon"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/response"
	"github.com/surlykke/RefudeServices/search"
)

//go:embed html
var sources embed.FS

var rowTemplate *template.Template
var StaticServer http.Handler

var funcMap = template.FuncMap{
	"tabindex": func(i, j int) int {
		if j == 0 {
			return i + 1
		} else {
			return -1
		}
	},
	"autofocus": func(i, j int) bool {
		return i == 0 && j == 0
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

	fmt.Println("Mapping SearchHandler")
	var tmp http.Handler

	if htmlDir, err := fs.Sub(sources, "html"); err == nil {
		tmp = http.FileServer(http.FS(htmlDir))
	} else {
		log.Panic(err)
	}
	StaticServer = http.StripPrefix("/desktop", tmp)

}

type Resourceline struct {
	Icon        icon.Name
	Title       string
	ActionLinks []Resourcelink
}

type Resourcelink struct {
	Relation  relation.Relation
	Href      string
	Icon      icon.Name
	Title     string
	Tabindex  int
	Autofocus string
}

func SearchHandler(term string) response.Response {
	var reslines = make([]Resourceline, 0, 50)
	for i, entity := range search.Search(term) {
		var resline = Resourceline{
			Icon:        entity.Icon,
			Title:       entity.Title,
			ActionLinks: make([]Resourcelink, 0, len(entity.Links))}
		for _, l := range entity.Links {
			if l.Relation != relation.Action {
				continue
			}
			var autofocus string
			if i == 0 && len(resline.ActionLinks) == 0 {
				autofocus = "autofocus"
			}
			var tabindex = -1
			if len(resline.ActionLinks) == 0 {
				tabindex = i + 1
			}
			resline.ActionLinks = append(resline.ActionLinks, Resourcelink{
				Relation:  l.Relation,
				Href:      l.Href,
				Icon:      l.Icon,
				Title:     l.Title,
				Tabindex:  tabindex,
				Autofocus: autofocus,
			})
		}
		reslines = append(reslines, resline)
	}

	var b bytes.Buffer
	if err := rowTemplate.Execute(&b, reslines); err != nil {
		return response.ServerError(err)
	} else {
		return response.HtmlBytes(b.Bytes())
	}
}
