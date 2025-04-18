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

func loadTemplate(name, relPath string) *template.Template {
	var bytes []byte
	var err error
	if bytes, err = sources.ReadFile(relPath); err != nil {
		log.Panic(err)
	}
	return template.Must(template.New(name).Parse(string(bytes)))
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
	var (
		reslines         = make([]Resourceline, 0, 50)
		nextTabIndex int = 1
	)

	for _, entity := range search.Search(term) {
		var resline = Resourceline{Icon: entity.Icon, Title: entity.Title, ActionLinks: make([]Resourcelink, 0, len(entity.Links))}

		for j, act := range entity.Actions {
			var autofocus string
			var tabindex = -1
			if j == entity.FocusHint {
				if nextTabIndex == 1 {
					autofocus = "autofocus"
				}
				tabindex = nextTabIndex
				nextTabIndex++
			}

			resline.ActionLinks = append(resline.ActionLinks, Resourcelink{
				Relation:  relation.Action,
				Href:      act.Href(entity.Path),
				Title:     act.Name,
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
		return response.Html(b.Bytes())
	}
}
