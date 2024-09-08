// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package file

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/icons"
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

func makeLinkFromPath(path string, name string) resource.Link {
	var title = name
	var mimetype, _ = magicmime.TypeByFile(path)
	var iconUrl = icons.UrlFromName(strings.ReplaceAll(mimetype, "/", "-"))
	return resource.Link{Href: "/file" + path, Title: title, IconUrl: iconUrl, Relation: relation.Related, Type: "application/vnd.org.refude.file+json"}
}

func makeFileFromInfo(osPath string, fileInfo os.FileInfo) *File {
	var fileType = getFileType(fileInfo.Mode())
	var comment = osPath
	var mimetype, _ = magicmime.TypeByFile(osPath)
	var iconUrl = icons.UrlFromName(strings.ReplaceAll(mimetype, "/", "-"))
	var f = File{
		ResourceData: *resource.MakeBase("/file/"+osPath[1:], fileInfo.Name(), comment, iconUrl, "file"),
		Type:         fileType,
		Permissions:  fileInfo.Mode().String(),
		Mimetype:     mimetype,
	}

	for i, app := range applications.GetHandlers(f.Mimetype) {
		f.apps = append(f.apps, app.DesktopId)
		if i == 0 {
			f.AddLink(f.Path+"?action="+app.DesktopId, "Open with "+app.Title, app.GetLinks().Get(relation.Icon).Href, relation.DefaultAction)
		} else {
			f.AddLink(f.Path+"?action="+app.DesktopId, "Open with "+app.Title, app.GetLinks().Get(relation.Icon).Href, relation.Action)
		}
	}

	if fileType == "Directory" {
		f.AddLink("/search?from="+f.Path, "", "", relation.Search)
	}

	return &f
}

func (f *File) Search(term string) resource.LinkList {
	var osPath = f.Path[len("/file"):]
	var collector = f.ResourceData.Search(term)
	var pathPattern = strings.Split(term, "/")

	if f.Type == "Directory" {
		collectFilesFromDir(&collector, osPath, pathPattern[0:len(pathPattern)-1])
	}
	return collector
}

func SearchDesktop(term string) resource.LinkList {
	var collector = make(resource.LinkList, 0, 100)
	collector = append(collector, makeLinkFromPath(xdg.Home, "Home"))
	var pathPattern = strings.Split(term, "/")
	pathPattern = pathPattern[0 : len(pathPattern)-1]
	collectFilesFromDirs(
		&collector,
		[]string{xdg.Home, xdg.ConfigHome, xdg.DownloadDir, xdg.DocumentsDir, xdg.MusicDir, xdg.VideosDir},
		pathPattern)

	return collector
}

func collectFilesFromDirs(sink *resource.LinkList, dirs []string, pathPattern []string) {
	for _, dir := range dirs {
		collectFilesFromDir(sink, dir, pathPattern)
	}
}

func collectFilesFromDir(sink *resource.LinkList, dir string, pathPattern []string) {
	fmt.Println("collectFilesFromDir, dir:", dir, "pathPattern:", pathPattern)
	if file, err := os.Open(dir); err != nil {
		log.Warn("Could not open", dir, err)
	} else if entries, err := file.ReadDir(-1); err != nil {
		log.Warn("Could not read", dir, err)
	} else {
		for _, entry := range entries {
			var name = entry.Name()
			var path = dir + "/" + name
			if len(pathPattern) > 0 {
				if len(pathPattern[0]) > 0 && entry.IsDir() && strings.Contains(strings.ToLower(name), pathPattern[0]) {
					collectFilesFromDir(sink, path, pathPattern[1:])
				}
			} else {
				fmt.Println("Collecting", path, name)
				*sink = append(*sink, makeLinkFromPath(path, name))
			}
		}
	}
}

func globSearch(collector *resource.RankedLinkList, dir string, term string) {
	// FIXME
}

func (f *File) DoPost(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File#DoPost", r.URL.Query())
	var appId = requests.GetSingleQueryParameter(r, "action", "")
	if appId == "" && len(f.apps) > 0 {
		appId = f.apps[0]
	}
	if appId == "" || !slices.Contains(f.apps, appId) {
		respond.NotFound(w)
	} else {
		if applications.OpenFile(appId, f.Path[5:]) {
			respond.Accepted(w)
		} else {
			respond.NotFound(w)
		}
	}
}
