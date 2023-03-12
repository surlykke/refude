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
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
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
	self        string
	FilePath    string
	Name        string
	Type        string
	Permissions string
	Mimetype    string
	Icon        string
	Apps        []string
}

func (f *File) Path() string {
	return f.FilePath[1:]
}

func (f *File) Presentation() (title string, comment string, icon link.Href, profile string) {
	return f.Name, f.FilePath, link.IconUrl(f.Icon), "file"
}

func (f *File) Links(self, term string) link.List {
	var ll = make(link.List, 0, 10)

	var rel = relation.DefaultAction
	for _, app := range applications.GetApps(f.Apps...) {
		if rnk := searchutils.Match(term, app.Name); rnk > -1 {
			ll = append(ll, link.Make(self+"?action="+app.DesktopId, "Open with "+app.Name, app.Icon, rel))
			rel = relation.Action
		}
	}

	if f.Type == "Directory" {
		ll = append(ll, SearchFrom(f.FilePath, term, "")...)
	}

	return ll
}

func (file *File) RelevantForSearch() bool {
	return true
}

// Assumes dir is a directory
func SearchFrom(dir, term, context string) link.List {
	var depth = len(term) / 3
	if depth > 2 {
		depth = 2
	}
	return searchRecursiveFrom(dir, term, context, depth)
}

func searchRecursiveFrom(dir, term, context string, depth int) link.List {
	var ll = make(link.List, 0, 30)
	var directoriesFound = make([]fs.DirEntry, 0, 10)

	if dirEntries, err := os.ReadDir(dir); err == nil {
		for _, dirEntry := range dirEntries {
			var entryPath = dir + "/" + dirEntry.Name()
			if rnk := searchutils.Match(term, dirEntry.Name()); rnk > -1 {
				var mimetype, _ = magicmime.TypeByFile(entryPath)
				var icon = strings.ReplaceAll(mimetype, "/", "-")
				ll = append(ll, link.MakeRanked(entryPath[1:], context+dirEntry.Name(), icon, "file", rnk+50))

			}
			if depth > 0 && dirEntry.IsDir() {
				directoriesFound = append(directoriesFound, dirEntry)
			}
		}
	}

	for _, directory := range directoriesFound {
		ll = append(ll, searchRecursiveFrom(dir+"/"+directory.Name(), term, context+directory.Name()+"/", depth-1)...)
	}
	return ll
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
			FilePath:    path,
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
	var ok, err = applications.OpenFile(appId, f.FilePath)
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
