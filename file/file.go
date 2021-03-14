package file

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type File struct {
	respond.Resource
	Path       string
	Mimetype   string
	DefaultApp string
}

func makeActor(filePath string, appId string) respond.Actor {
	return func(*http.Request) error {
		if app := applications.GetApp(appId); app != nil {
			app.Run(filePath)
			return nil
		} else {
			return fmt.Errorf("No such app: '%s'", appId)
		}
	}
}

func makeFile(path string) (*File, error) {
	if !strings.HasPrefix(path, "/") {
		path = xdg.Home + "/" + path
	}

	var hasher = sha1.New()
	hasher.Write([]byte(path))
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else {
		var self = "/file?path=" + url.QueryEscape(path)
		var mimetype, _ = magicmime.TypeByFile(path)
		var f = File{
			Path:       path,
			Mimetype:   mimetype,
			DefaultApp: applications.GetDefaultApp(mimetype),
		}
		f.Resource = respond.MakeResource(self, path, applications.IconForMimetype(mimetype), &f, "file")

		var recommendedApps, _ = applications.GetAppsForMimetype(f.Mimetype)
		for i, app := range recommendedApps {
			var actionId = ""
			if i != 0 {
				actionId = app.Id
			}
			f.AddAction(respond.MakeAction(actionId, "Open with "+app.Name, applications.Icon2IconUrl(app.Icon), makeActor(path, app.Id)))
		}

		return &f, nil
	}
}
