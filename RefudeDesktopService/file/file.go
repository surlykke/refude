package file

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/applications"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type File struct {
	respond.Links `json:"_links"`
	Path          string
	Mimetype      string
	DefaultApp    string
}

func (f *File) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, f)
	} else if r.Method == "POST" {
		if appId := requests.GetSingleQueryParameter(r, "appid", ""); appId != "" {
			if app := applications.GetApp(appId); app != nil {
				respond.AcceptedAndThen(w, func() { app.Run(f.Path) })
			} else {
				respond.NotFound(w)
			}
		} else {
			if err := applications.OpenFile(f.Path, f.Mimetype); err != nil {
				respond.ServerError(w, err)
			} else {
				respond.Accepted(w)
			}
		}

	} else {
		respond.NotAllowed(w)
	}
}

func makeFile(path string) (*File, error) {
	if !strings.HasPrefix(path, "/") {
		path = xdg.Home + "/" + path
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else {
		var self = "/file?path=" + url.QueryEscape(path)
		var mimetype, _ = magicmime.TypeByFile(path)
		var f = File{
			Links: respond.Links{{
				Href:    self,
				Rel:     respond.Self,
				Title:   path,
				Profile: "/profile/file",
				Icon:    applications.IconForMimetype(mimetype),
			}},
			Path:       path,
			Mimetype:   mimetype,
			DefaultApp: applications.GetDefaultApp(mimetype),
		}

		var recommendedApps, _ = applications.GetAppsForMimetype(f.Mimetype)
		for _, app := range recommendedApps {
			f.Links = append(f.Links, respond.Link{
				Href:  self + "&appid=" + app.Id,
				Title: "Open with " + app.Name,
				Icon:  applications.Icon2IconUrl(app.Icon),
				Rel:   respond.Action,
			})
		}

		return &f, nil
	}
}
