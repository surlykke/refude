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
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/wayland"
)

//go:embed html
var sources embed.FS

var resourceTemplate *template.Template
var rowTemplate *template.Template
var StaticServer http.Handler

var funcMap = template.FuncMap{
	// The name "inc" is what the function will be called in the template text.
	"inc": func(i int) int {
		return i + 1
	},
	"comment": func(l resource.Link) string {
		switch l.Type {
		case mediatype.Window, mediatype.Tab:
			return "Focus"
		case mediatype.Application:
			return "Launch"
		default:
			return ""
		}
	},
}

func init() {
	var bytes []byte
	var err error

	if bytes, err = sources.ReadFile("html/resourceTemplate.html"); err != nil {
		log.Panic(err)
	}
	resourceTemplate = template.Must(template.New("resourceTemplate").Parse(string(bytes)))

	if bytes, err = sources.ReadFile("html/rowTemplate.html"); err != nil {
		log.Panic(err)
	}
	rowTemplate = template.Must(template.New("rowTemplate").Funcs(funcMap).Parse(string(bytes)))

}

type item struct {
	IconUrl string
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
	case "/desktop/resource":
		if r.Method != "GET" {
			respond.NotAllowed(w)
		} else if res := getResource(r); res == nil {
			respond.NotFound(w)
		} else {
			var m = map[string]any{
				"Title": res.GetTitle(),
				"Icon":  res.GetLinks().Get(relation.Icon).Href,
				"Path":  res.GetPath(),
			}
			m["Data"] = defaultData(res)
			if err := resourceTemplate.Execute(w, m); err != nil {
				log.Warn("Error executing resourceTemplate:", err)
			}

		}
	case "/desktop/search":
		if r.Method != "GET" {
			respond.NotAllowed(w)
		} else if res := getResource(r); res == nil {
			respond.NotFound(w)
		} else {
			var term = strings.ToLower(requests.GetSingleQueryParameter(r, "search", ""))
			var m = map[string]any{
				"Term":  term,
				"Links": res.Search(term).FilterAndSort(term),
			}
			if err := rowTemplate.Execute(w, m); err != nil {
				log.Warn("Error executing rowTemplate:", err)
			}
		}
	/* FIXME case "/desktop/tray":

	if r.Method != "GET" {
		respond.NotAllowed(w)
	} else {
		var items = make([]item, 0, 10)
		for _, i := range repo.GetListSortedByPath[*statusnotifications.Item]("/item/") {
			items = append(items, item{IconUrl: i.GetIconUrl()})
		}
		if err := bodyTemplate.Execute(w, map[string]any{"Items": items}); err != nil {
			log.Warn("Error executing bodyTemplate:", err)
		}

	}*/
	default:
		if strings.HasSuffix(r.URL.Path, "Template.html") {
			respond.NotFound(w)
		} else {
			StaticServer.ServeHTTP(w, r)
		}
	}
}

func getResource(r *http.Request) resource.Resource {
	var resourcePath = strings.Replace(requests.GetSingleQueryParameter(r, "path", "/start"), "http://localhost:7938", "", 1)
	if strings.HasPrefix(resourcePath, "/file/") {
		return file.GetResource(resourcePath)
	} else {
		return repo.GetUntyped(resourcePath)
	}
}

func defaultData(res resource.Resource) [][]string {

	switch res.GetType() {
	case "window":
		var window = res.(*wayland.WaylandWindow)
		return [][]string{
			{"Wid", fmt.Sprintf("%d", window.Wid)},
			{"AppId", window.AppId},
			{"State", window.State.String()},
		}
	case "application":
		var application = res.(*applications.DesktopApplication)
		return [][]string{
			{"Type", application.Type},
			{"Version", application.Version},
			{"GenericName", application.GenericName},
			{"NoDisplay", showBool(application.NoDisplay)},
			{"Exec", application.Exec},
			{"Terminal", showBool(application.Terminal)},
			{"Categories", strings.Join(application.Categories, ", ")},
			{"DesktopId", application.DesktopId},
			{"Mimetypes", strings.Join(application.Mimetypes, ", ")},
			{"DesktopFile", application.DesktopFile},
		}
	case "device":
		var dev = res.(*power.Device)
		return [][]string{
			{"Energy", fmt.Sprintf("%f", dev.Energy)},
			{"EnergyEmpty", fmt.Sprintf("%f", dev.EnergyEmpty)},
			{"EnergyFull", fmt.Sprintf("%f", dev.EnergyFull)},
			{"EnergyFullDesign", fmt.Sprintf("%f", dev.EnergyFullDesign)},
			{"EnergyRate", fmt.Sprintf("%f", dev.EnergyRate)},
			{"Percentage", fmt.Sprintf("%d", dev.Percentage)},
			{"TimeToEmpty", fmt.Sprintf("%d", dev.TimeToEmpty)},
			{"TimeToFull", fmt.Sprintf("%d", dev.TimeToFull)},
			{"DisplayDevice", showBool(dev.DisplayDevice)},
			{"NativePath", dev.NativePath},
			{"Vendor", dev.Vendor},
			{"Model", dev.Model},
			{"Serial", dev.Serial},
			{"UpdateTime", fmt.Sprintf("%d", dev.UpdateTime)},
			{"Type", dev.Type},
			{"PowerSupply", showBool(dev.PowerSupply)},
			{"Online", showBool(dev.Online)},
			{"Voltage", fmt.Sprintf("%f", dev.Voltage)},
			{"IsPresent", showBool(dev.IsPresent)},
			{"State", dev.State},
			{"IsRechargeable", showBool(dev.IsRechargeable)},
			{"Capacity", fmt.Sprintf("%f", dev.Capacity)},
			{"Technology", dev.Technology},
			{"Warninglevel", dev.Warninglevel},
			{"Batterylevel", dev.Batterylevel},
		}

	default:
		return [][]string{}
	}
}

func showBool(b bool) string {
	if b {
		return "yes"
	} else {
		return "no"
	}
}
