package file

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

func init() {
	if err := magicmime.Open(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR); err != nil {
		panic(err)
	}
}

type File struct {
	Path       string
	self       string
	Name       string
	Dir        bool
	Mimetype   string
	Icon       string
	DefaultApp string
}

func makeFile(path string) (*File, error) {
	if !strings.HasPrefix(path, "/") {
		path = xdg.Home + "/" + path
	}
	path = filepath.Clean(path)

	if fileInfo, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else {
		var mimetype, _ = magicmime.TypeByFile(path)
		var f = File{
			Path:       path,
			self:       "/file/" + url.PathEscape(path),
			Name:       fileInfo.Name(),
			Dir:        fileInfo.IsDir(),
			Mimetype:   mimetype,
			Icon:       strings.ReplaceAll(mimetype, "/", "-"),
			DefaultApp: applications.GetDefaultApp(mimetype),
		}

		return &f, nil
	}
}

func (f *File) Links() []resource.Link {
	var links = []resource.Link{resource.MakeLink(f.self, f.Name, f.Icon, relation.Self)}

	var recommendedApps, _ = applications.GetAppsForMimetype(f.Mimetype)
	for i, app := range recommendedApps {
		if i == 0 {
			links = append(links, resource.MakeLink(f.self, "Open with "+app.Name, app.Icon, relation.DefaultAction))
		} else {
			links = append(links, resource.MakeLink(f.self+"?action="+app.Id, "Open with "+app.Name, app.Icon, relation.Action))
		}
	}

	if f.Dir {
		if dir, err := os.Open(f.Path); err != nil {
			log.Warn("Error opening", f.Path, err)
		} else if names, err := dir.Readdirnames(-1); err != nil {
			log.Warn("Error reading", f.Path, err)
			dir.Close()
		} else {
			// Can't use filepath.Glob as it is case sensitive
			for _, name := range names {
				var path = f.Path + "/" + name
				var mimetype, _ = magicmime.TypeByFile(path)
				var icon = strings.ReplaceAll(mimetype, "/", "-")
				links = append(links, resource.MakeLink("/file/"+url.PathEscape(path), name, icon, relation.Related))
			}
			dir.Close()
		}

	}

	return links
}

func (f *File) RefudeType() string {
	return "file"
}

func (f *File) DoPost(w http.ResponseWriter, r *http.Request) {
	var appId = requests.GetSingleQueryParameter(r, "action", "")
	if appId == "" {
		applications.OpenFile(f.Path, f.Mimetype)
		respond.Accepted(w)
	} else if app := applications.GetApp(appId); app != nil {
		if err := app.Run(f.Path); err != nil {
			respond.ServerError(w, err)
		} else {
			respond.Accepted(w)
		}
	} else {
		respond.NotFound(w)
	}

}
