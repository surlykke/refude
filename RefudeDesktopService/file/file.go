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
	Path       string
	Mimetype   string
	DefaultApp string
}

func (f *File) ToStandardFormat() *respond.StandardFormat {
	var self = "/file?path=" + url.QueryEscape(f.Path)

	var comment = "Open"
	if f.DefaultApp != "" {
		comment += " with " + f.DefaultApp
	}

	var Actions = make([]respond.Action, 0, 10)
	var recommendedApps, _ = applications.GetAppsForMimetype(f.Mimetype)

	for _, app := range recommendedApps {
		Actions = append(Actions, respond.Action{
			Title:    app.Name,
			IconName: app.IconName,
			Path:     self + "&appid=" + app.Id,
		})
	}

	return &respond.StandardFormat{
		Self:     self,
		Type:     "file",
		Title:    f.Path,
		Comment:  comment,
		IconName: strings.ReplaceAll(f.Mimetype, "/", "-"),
		OnPost:   "Open",
		Actions:  Actions,
		Data:     f,
	}
}

func (f *File) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, f.ToStandardFormat())
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
		var mimetype, _ = magicmime.TypeByFile(path)
		return &File{
			Path:       path,
			Mimetype:   mimetype,
			DefaultApp: applications.GetDefaultAppName(mimetype),
		}, nil
	}
}
