// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package desktop

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/icon"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/response"
	"github.com/surlykke/RefudeServices/search"
)

//go:embed html
var sources embed.FS

var rowTemplate *template.Template
var detailsTemplate *template.Template

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
	detailsTemplate = loadTemplate("detailsTemplate", "html/detailsTemplate.html")

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
	Comment     string
	Href        string
	Path        string
	MoreActions bool
}

func SearchHandler(term string) response.Response {
	var (
		lines []Resourceline
	)

	for _, r := range search.Search(term) {

		var line = Resourceline{Icon: r.Icon, Title: r.Title, Comment: r.Subtitle}
		if len(r.Actions) > 0 {
			line.Href = r.Actions[0].Href(r.Path)
			line.Path = r.Path
		}
		line.MoreActions = len(r.Actions) > 1
		lines = append(lines, line)
	}

	var b bytes.Buffer
	if err := rowTemplate.Execute(&b, lines); err != nil {
		log.Error(err)
		return response.ServerError(err)
	} else {
		return response.Html(b.Bytes())
	}
}

type Detail struct {
	Name string
	Href string
}

func DetailsHandler(resPath string) response.Response {
	var b bytes.Buffer
	if base, ok := search.SearchByPath(resPath); !ok {
		return response.NotFound()
	} else if err := detailsTemplate.Execute(&b, base.ActionLinks()); err != nil {
		log.Error(err)
		return response.ServerError(err)
	} else {
		return response.Html(b.Bytes())
	}
}
