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
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
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

func makeFileFromInfo(osPath string, fileInfo os.FileInfo) *File {
	var mimetype, _ = magicmime.TypeByFile(osPath)
	var fileType = getFileType(fileInfo.Mode())
	var f = File{
		BaseResource: *resource.MakeBase("/file/"+osPath[1:], fileInfo.Name(), osPath, strings.ReplaceAll(mimetype, "/", "-"), "file"),
		Type:         fileType,
		Permissions:  fileInfo.Mode().String(),
		Mimetype:     mimetype,
		Apps:         applications.GetAppsIds(mimetype),
	}

    if fileType == "Directory" {
		f.AddLink("/search?from=" + f.Path, "", "", relation.Search)
	}	

	for _, app := range applications.GetApps(f.Apps...) {
		f.AddLink("?action=" + app.DesktopId, "Open with " + app.Title, app.IconUrl, relation.Action)
	}

	return &f
}

func (f *File) Search(searchTerm string) []resource.Resource {
	return Search(f.Path[len("/file"):], ".", searchTerm)
}

// Assumes dir is a directory
func Search(from, prefix, searchTerm string) []resource.Resource {
	var depth = len(searchTerm) / 3
	if depth > 2 {
		depth = 2
	}
	return searchRecursive(from, prefix, searchTerm, depth)
}

func searchRecursive(from, prefix, searchTerm string, depth int) []resource.Resource {
	var fileList = make([]resource.Resource, 0, 30)
	var directoriesFound = make([]fs.DirEntry, 0, 10)

	if dirEntries, err := os.ReadDir(from); err == nil {
		for _, dirEntry := range dirEntries {
			var entryPath = from + "/" + dirEntry.Name()
			var relName = dirEntry.Name()
			if prefix != "." {
				relName = prefix + "/" + relName
			}
			if rnk := searchutils.Match(searchTerm, dirEntry.Name()); rnk > -1 {
				if info, err := dirEntry.Info(); err != nil {
					log.Warn(err)
				} else {
					fileList = append(fileList, makeFileFromInfo(entryPath, info))
				}
			}
			if depth > 0 && dirEntry.IsDir() {
				directoriesFound = append(directoriesFound, dirEntry)
			}
		}
	}

	for _, directory := range directoriesFound {
		fileList = append(fileList, searchRecursive(from+"/"+directory.Name(), prefix+"/"+directory.Name(), searchTerm, depth-1)...)
	}
	return fileList
}

func (f *File) DoPost(w http.ResponseWriter, r *http.Request) {
	var defaultAppId = ""
	if len(f.Apps) > 0 {
		defaultAppId = f.Apps[0]
	}
	var appId = requests.GetSingleQueryParameter(r, "action", defaultAppId)
	var ok, err = applications.OpenFile(appId, f.Path[len("/file"):])
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
