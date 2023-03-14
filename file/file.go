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
	resource.BaseResource
	Name        string
	Type        string
	Permissions string
	Mimetype    string
	Apps        []string
}

func makeFile(path string) (*File, error) {
	var osPath = filepath.Clean("/" + path)	
	if fileInfo, err := os.Stat(osPath); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else {
		var mimetype, _ = magicmime.TypeByFile(osPath)
		var f = File{
			BaseResource: resource.BaseResource{
				Path:     osPath[1:],
				Title:    fileInfo.Name(),
				Comment:  path,
				IconName: strings.ReplaceAll(mimetype, "/", "-"),
			},
			Type:        getFileType(fileInfo.Mode()),
			Permissions: fileInfo.Mode().String(),
			Mimetype:    mimetype,
			Apps:        applications.GetAppsIds(mimetype),
		}

		return &f, nil
	}
}


func (f *File) Actions() link.ActionList {
	var actions = make(link.ActionList, 0, 10)
	for _, app := range applications.GetApps(f.Apps...) {
		actions = append(actions, link.MkAction(app.DesktopId, "Open with " + app.Title, app.IconName))
	}
	return actions
}

func (f *File) Links(searchTerm string) link.List {

	var ll = make(link.List, 0, 10)
	if f.Type == "Directory" {
		ll = append(ll, Search("/" + f.Path, searchTerm)...)
	}

	return ll
}

// Assumes dir is a directory
func Search(from, searchTerm string) link.List {
	var depth = len(searchTerm) / 3
	if depth > 2 {
		depth = 2
	}
	return searchRecursive(from, searchTerm, depth)
}

func searchRecursive(from, searchTerm string, depth int) link.List {
	var ll = make(link.List, 0, 30)
	var directoriesFound = make([]fs.DirEntry, 0, 10)

	if dirEntries, err := os.ReadDir(from); err == nil {
		for _, dirEntry := range dirEntries {
			var entryPath = from + "/" + dirEntry.Name()
			if rnk := searchutils.Match(searchTerm, dirEntry.Name()); rnk > -1 {
				var mimetype, _ = magicmime.TypeByFile(entryPath)
				var icon = strings.ReplaceAll(mimetype, "/", "-")
				ll = append(ll, link.MakeRanked(entryPath[1:], shortenPath(entryPath), icon, "file", rnk+50))

			}
			if depth > 0 && dirEntry.IsDir() {
				directoriesFound = append(directoriesFound, dirEntry)
			}
		}
	}

	for _, directory := range directoriesFound {
		ll = append(ll, searchRecursive(from+"/"+directory.Name(), searchTerm, depth-1)...)
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

func (f *File) DoPost(w http.ResponseWriter, r *http.Request) {
	var defaultAppId = ""
	if len(f.Apps) > 0 {
		defaultAppId = f.Apps[0]
	}
	var appId = requests.GetSingleQueryParameter(r, "action", defaultAppId)
	var ok, err = applications.OpenFile(appId, "/" + f.Path)
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
