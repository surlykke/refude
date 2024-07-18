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
	"reflect"
	"strings"

	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/stringhash"
	"github.com/surlykke/RefudeServices/watch"
	"github.com/surlykke/RefudeServices/wayland"
)

//go:embed html
var sources embed.FS

var bodyTemplate *template.Template
var StaticServer http.Handler

func init() {
	var bytes []byte
	var err error

	if bytes, err = sources.ReadFile("html/bodyTemplate.html"); err != nil {
		log.Panic(err)
	}

	bodyTemplate = template.Must(template.New("bodyTemplate").Parse(string(bytes)))
}

type row struct {
	//	Heading  string
	Class    string
	IconUrl  string
	Title    string
	Comment  string
	Href     string
	Relation relation.Relation
	Profile  string
}

func headingRow(heading string) row {
	return row{Title: heading, Class: "heading"}
}

func actionRow(action link.Link) row {
	return row{IconUrl: action.IconUrl, Title: action.Title, Href: action.Href, Relation: action.Relation, Class: "selectable"}
}

func resourceRow(sr resource.Resource) row {
	var comment string
	if sr.GetComment() != "" {
		comment = sr.GetProfile() + ": " + sr.GetComment()
	}
	return row{IconUrl: sr.GetIconUrl(), Title: sr.GetTitle(), Comment: comment, Href: sr.GetPath(), Relation: relation.Self, Profile: sr.GetProfile(), Class: "selectable"}
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
	case "/desktop/body":
		if r.Method != "GET" {
			respond.NotAllowed(w)
		} else {
			var resourcePath = requests.GetSingleQueryParameter(r, "resource", "/start")
			var res resource.Resource = nil
			if strings.HasPrefix(resourcePath, "/file/") {
				res = file.GetResource(resourcePath)
			} else {
				res = repo.GetUntyped(resourcePath)
			}

			if res != nil {
				var (
					term           = strings.ToLower(requests.GetSingleQueryParameter(r, "search", ""))
					actions        = res.GetActionLinks(term)
					sf, searchable = res.(resource.Searchable)
					rows           = make([]row, 0, len(actions)+4)
				)
				if len(actions) > 0 {
					rows = append(rows, headingRow("Actions"))
				}
				for _, a := range actions {
					rows = append(rows, actionRow(a))
				}
				if len(actions) > 0 {
					rows = append(rows, headingRow("Related"))
				}
				if searchable {
					for _, subresGroup := range arrange(sf.Search(term)) {
						if len(subresGroup.resources) > 0 {
							rows = append(rows, headingRow(subresGroup.heading))
							for _, subres := range subresGroup.resources {
								rows = append(rows, resourceRow(subres))
							}
						}
					}
				}

				var m = map[string]any{
					"Searchable": searchable,
					"Title":      res.GetTitle(),
					"Icon":       res.GetIconUrl(),
					"Term":       term,
					"Rows":       rows,
				}
				var etag = buildETag(term, res.GetTitle(), res.GetIconUrl(), rows)
				if r.Header.Get("if-none-match") == etag {
					respond.NotModified(w)
					return
				}
				w.Header().Set("ETag", etag)
				if err := bodyTemplate.Execute(w, m); err != nil {
					log.Warn("Error executing bodyTemplate:", err)
				}

			} else {
				respond.NotFound(w)
			}
		}
	case "/desktop/show":
		if r.Method != "POST" {
			respond.NotAllowed(w)
		} else {
			wayland.RememberActive()
			watch.Publish("showDesktop", "")
			respond.Accepted(w)
		}
	case "/desktop/hash":
		if r.Method == "GET" {
			// Go Json cannot handle uint64, so we convert to string
			respond.AsJson(w, "")
		} else {
			respond.NotAllowed(w)
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
			respond.Accepted(w)
		}
	case "/desktop/bodyTemplate.html":
		respond.NotFound(w)
	default:
		StaticServer.ServeHTTP(w, r)
	}
}

func defaultData(res resource.Resource) string {
	var structAsMap = make(map[string]interface{})
	var Type = reflect.TypeOf(res)
	if Type.Kind() != reflect.Struct {
		panic("Resource not a struct")
	}
	var structVal = reflect.ValueOf(res)
outer:
	for i := 0; i < Type.NumField(); i++ {
		if Type.Field(i).IsExported() {
			var fieldName = Type.Field(i).Name
			for _, omitted := range []string{"Title", "Comment"} {
				if fieldName == omitted {
					continue outer
				}
			}
			structAsMap[fieldName] = structVal.Field(i)
		}
	}
	return string(respond.ToJson(structAsMap))
}

func buildETag(term string, title, icon string, rows []row) string {
	var hash uint64 = 0
	hash = stringhash.FNV1a(term, title, icon)
	for _, row := range rows {
		hash = hash ^ stringhash.FNV1a(row.Title, row.Comment, row.Href, row.IconUrl, row.Profile, string(row.Relation))
	}
	return fmt.Sprintf(`"%X"`, hash)
}

type resourceGroup struct {
	heading   string
	resources []resource.Resource
}

func arrange(resources []resource.Resource) []resourceGroup {
	var groups = []resourceGroup{{"Notifications", nil}, {"Windows and Tabs", nil}, {"Applications", nil}, {"Files", nil}, {"Devices", nil}, {"Other", nil}}
	for _, res := range resources {
		if res.GetProfile() == "notification" {
			groups[0].resources = append(groups[0].resources, res)
		} else if res.GetProfile() == "window" || res.GetProfile() == "tab" {
			groups[1].resources = append(groups[1].resources, res)
		} else if res.GetProfile() == "application" {
			groups[2].resources = append(groups[2].resources, res)
		} else if res.GetProfile() == "file" {
			groups[3].resources = append(groups[3].resources, res)
		} else if res.GetProfile() == "device" {
			groups[4].resources = append(groups[4].resources, res)
		} else {
			groups[5].resources = append(groups[5].resources, res)
		}
	}
	return groups
}
