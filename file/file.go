package file

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

func init() {
	if err := magicmime.Open(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR); err != nil {
		panic(err)
	}
}

func getFileType(m os.FileMode) string {
	if m&fs.ModeDir > 0 {
		return "Directory"
	} else if m&fs.ModeSymlink > 0 {
		return "Symbolic link"
	} else if m&fs.ModeNamedPipe > 0 {
		return "Named pipe"
	} else if m&fs.ModeSocket > 0 {
		return "Socket"
	} else if m&fs.ModeDevice > 0 {
		return "Device"
	} else if m&fs.ModeCharDevice > 0 {
		return "Char device"
	} else if m&fs.ModeIrregular > 0 {
		return "Irregular"
	} else {
		return "File"
	}
}

type File struct {
	Path        string
	self        string
	Name        string
	Type        string
	Permissions string
	Mimetype    string
	Icon        string
	Apps        []string
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
			Path:        path,
			self:        "/file" + path,
			Name:        fileInfo.Name(),
			Type:        getFileType(fileInfo.Mode()),
			Permissions: fileInfo.Mode().String(),
			Mimetype:    mimetype,
			Icon:        strings.ReplaceAll(mimetype, "/", "-"),
			Apps:        applications.GetAppsIds(mimetype),
		}

		return &f, nil
	}
}

func (f *File) IsSearchable() bool {
	return f.Type == "Directory"
}

func (f *File) GetLinks(term string) link.List {
	var ll = make(link.List, 0, 10)

	if f.Type == "Directory" {

		if candidatePaths, err := filepath.Glob(f.Path + "/*"); err == nil { // TODO: readdir faster?
			for _, path := range candidatePaths {
				var fileName = filepath.Base(path)
				// Hidden files should be a little harder to find
				if strings.HasPrefix(fileName, ".") && !strings.HasPrefix(term, ".") {
					continue
				}
				if rnk := searchutils.Match(term, fileName); rnk > -1 {
					var mimetype, _ = magicmime.TypeByFile(path)
					var icon = strings.ReplaceAll(mimetype, "/", "-")
					ll = append(ll, link.MakeRanked("/file"+path, shortenPath(path), icon, "file", rnk+50))
				}
			}
		}

	}

	return ll
}

func shortenPath(path string) string {
	if strings.HasPrefix(path, xdg.Home) {
		return "~" + path[len(xdg.Home):]
	} else {
		return path
	}
}

func (f *File) GetPostActions() []resource.Action {
	var actions = make([]resource.Action, 0, 10)
	for _, app := range applications.GetApps(f.Apps...) {
		actions = append(actions, resource.Action{Id: app.Id, Title: "Open with " + app.Name, Icon: app.Icon})
	}
	return actions
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
