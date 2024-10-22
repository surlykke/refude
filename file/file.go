// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package file

import (
	"io/fs"
	"net/http"
	"os"
	gopath "path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/icon"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/path"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
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
	resource.ResourceData
	Name        string
	Type        string
	Permissions string
	Mimetype    string
	apps        []string
}

func makeFileFromPath(path string) (*File, error) {
	var osPath = filepath.Clean("/" + path)
	if fileInfo, err := os.Stat(osPath); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else {
		return makeFileFromInfo(osPath, fileInfo), nil
	}
}

func MakeLinkFromPath(ospath string, name string) resource.Link {
	var title = name
	var mimetype, _ = magicmime.TypeByFile(ospath)
	var icon = icon.Name(strings.ReplaceAll(mimetype, "/", "-"))
	return resource.Link{Path: path.Of("/file", gopath.Clean(ospath)), Title: title, Icon: icon, Relation: relation.Related, Type: "application/vnd.org.refude.file+json"}
}

func makeFileFromInfo(osPath string, fileInfo os.FileInfo) *File {
	var fileType = getFileType(fileInfo.Mode())
	var comment = osPath
	var mimetype, _ = magicmime.TypeByFile(osPath)
	var icon = icon.Name(strings.ReplaceAll(mimetype, "/", "-"))
	var path = path.Of("/file" + gopath.Clean(osPath))
	var f = File{
		ResourceData: *resource.MakeBase(path, fileInfo.Name(), comment, icon, mediatype.File),
		Type:         fileType,
		Permissions:  fileInfo.Mode().String(),
		Mimetype:     mimetype,
	}

	for i, app := range applications.GetHandlers(f.Mimetype) {
		f.Actions = make([]resource.Action, 0, 5)
		if i == 0 {
			f.DefaultAction = "Open with " + app.Title
		} else {
			f.Actions = append(f.Actions, resource.Action{Id: app.DesktopId, Title: "Open with " + app.Title, Icon: app.Icon})
		}
	}

	return &f
}

func Collector(dirs []string) func(string) []resource.Link {
	return func(string) []resource.Link {
		var result = make([]resource.Link, 0, 50)
		for _, dir := range dirs {
			for _, entry := range readEntries(dir) {
				var name = entry.Name()
				result = append(result, MakeLinkFromPath(dir+"/"+name, name))
			}
		}
		return result
	}
}

func Collect(sink *[]resource.Link, dir string) {
	for _, entry := range readEntries(dir) {
		var name = entry.Name()
		*sink = append(*sink, MakeLinkFromPath(dir+"/"+name, name))
	}
}

func CollectDirs(dirs []string, pathBits []string) []string {
	var collected = make([]string, 0, 30)
	for _, dir := range dirs {
		collectDirs2(&collected, dir, pathBits)
	}
	return collected
}

func collectDirs2(sink *[]string, dir string, pathBits []string) {
	if len(pathBits) == 0 {
		return
	}
	for _, entry := range readEntries(dir) {
		var name = entry.Name()
		var path = dir + "/" + name
		if entry.IsDir() && strings.Contains(strings.ToLower(name), pathBits[0]) {
			if len(pathBits) == 1 {
				*sink = append(*sink, path)
			} else {
				collectDirs2(sink, path, pathBits[1:])
			}
		}
	}
}

func readEntries(dir string) []fs.DirEntry {
	if file, err := os.Open(dir); err != nil {
		log.Warn("Could not open", dir, err)
		return nil
	} else if entries, err := file.ReadDir(-1); err != nil {
		log.Warn("Could not read", dir, err)
		return nil
	} else {
		return entries
	}
}

func (f *File) DoPost(w http.ResponseWriter, r *http.Request) {
	var appId = requests.GetSingleQueryParameter(r, "action", "")
	if appId == "" && len(f.apps) > 0 {
		appId = f.apps[0]
	}
	if appId == "" || !slices.Contains(f.apps, appId) {
		respond.NotFound(w)
	} else {
		var ospath = string(f.Path[5:])
		if applications.OpenFile(appId, ospath) {
			respond.Accepted(w)
		} else {
			respond.NotFound(w)
		}
	}
}
