package file

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

func init() {
	if err := magicmime.Open(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR); err != nil {
		panic(err)
	}
}

type File struct {
	respond.Resource
	Path       string
	Mimetype   string
	DefaultApp string
}

func makeFile(path string) (*File, error) {
	if !strings.HasPrefix(path, "/") {
		path = xdg.Home + "/" + path
	}

	if fileInfo, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else {
		var mimetype, _ = magicmime.TypeByFile(path)
		var f = File{
			Resource:   makeResource(path, mimetype),
			Path:       path,
			Mimetype:   mimetype,
			DefaultApp: applications.GetDefaultApp(mimetype),
		}

		var recommendedApps, _ = applications.GetAppsForMimetype(f.Mimetype)
		for i, app := range recommendedApps {
			if i == 0 {
				f.AddDefaultActionLink("Open with "+app.Name, applications.Icon2IconUrl(app.Icon))
			} else {
				f.AddActionLink("Open with "+app.Name, applications.Icon2IconUrl(app.Icon), app.Id)
			}
		}

		if fileInfo.IsDir() {
			if dir, err := os.Open(path); err != nil {
				log.Warn("Error opening", path, err)
			} else if names, err := dir.Readdirnames(-1); err != nil {
				log.Warn("Error reading", path, err)
				dir.Close()
			} else {
				// Can't use filepath.Glob as it is case sensitive
				for _, name := range names {
					var path = path + "/" + name
					var mimetype, _ = magicmime.TypeByFile(path)
					var resource = makeResource(path, mimetype)
					f.Links = append(f.Links, resource.GetRelatedLink())
				}
				dir.Close()
			}

		}

		return &f, nil
	}
}

func makeResource(path, mimetype string) respond.Resource {
	var title = path
	if strings.HasPrefix(title, xdg.Home) {
		title = "~" + title[len(xdg.Home):]
	}
	return respond.MakeResource("/file?path="+url.QueryEscape(path), title, applications.IconForMimetype(mimetype), "file")
}

func (f *File) DoPost(w http.ResponseWriter, r *http.Request) {
	var appId = requests.GetSingleQueryParameter(r, "action", "")
	if appId == "" {
		appId = f.DefaultApp
	}
	if app := applications.GetApp(appId); app != nil {
		if err := app.Run(f.Path); err != nil {
			respond.ServerError(w, err)
		} else {
			respond.Accepted(w)
		}
	} else {
		respond.NotFound(w)
	}

}
