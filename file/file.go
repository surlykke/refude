package file

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/relation"
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
	Path     string
	self     string
	Name     string
	Dir      bool
	Mimetype string
	Icon     string
	Apps     []string
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
			Path:     path,
			self:     "/file" + path,
			Name:     fileInfo.Name(),
			Dir:      fileInfo.IsDir(),
			Mimetype: mimetype,
			Icon:     strings.ReplaceAll(mimetype, "/", "-"),
			Apps:     applications.GetAppsIds(mimetype),
		}

		return &f, nil
	}
}

func (f *File) Links(path string) link.List {
	var ll = make(link.List, 0, 10)
	for i, app := range applications.GetApps(f.Apps...) {
		if i == 0 {
			ll = ll.Add(path+"?action="+app.Id, "Open with "+app.Name, app.Icon, relation.DefaultAction)
		} else {
			ll = ll.Add(path+"?action="+app.Id, "Open with "+app.Name, app.Icon, relation.Action)
		}
	}

	if f.Dir {
		if dir, err := os.Open(f.Path); err != nil {
			log.Warn("Error opening", f.Path, err)
		} else if names, err := dir.Readdirnames(-1); err != nil {
			log.Warn("Error reading", f.Path, err)
			dir.Close()
		} else {
			for _, name := range names {
				var path = f.Path + "/" + name
				var mimetype, _ = magicmime.TypeByFile(path)
				var icon = strings.ReplaceAll(mimetype, "/", "-")
				ll = ll.Add("/file/"+url.PathEscape(path), name, icon, relation.Related)
			}
			dir.Close()
		}

	}

	return ll
}

func (f *File) ForDisplay() bool {
	return true
}

func (f *File) DoPost(w http.ResponseWriter, r *http.Request) {
	var defaultAppId = ""
	if len(f.Apps) > 0 {
		defaultAppId = f.Apps[0]
	}
	var appId = requests.GetSingleQueryParameter(r, "action", defaultAppId)
	var ok, err = applications.OpenFile(appId, f.Path)
	if ok {
		if err != nil {
			respond.ServerError(w, err)
		} else {
			respond.Accepted(w)
		}
	} else {
		respond.NotFound(w)
	}
}
