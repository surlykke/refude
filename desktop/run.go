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
	"github.com/surlykke/RefudeServices/lib/path"
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

type item struct {
	Icon     icon.Name
	ItemPath path.Path
	MenuPath path.Path
}

func SearchHandler(term string) response.Response {
	var m = map[string]any{
		"term":      term,
		"resources": search.Search(term),
	}
	var b bytes.Buffer
	if err := rowTemplate.Execute(&b, m); err != nil {
		return response.ServerError(err)
	} else {
		return response.HtmlBytes(b.Bytes())
	}
}
