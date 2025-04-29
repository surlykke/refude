// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package file

import (
	"io/fs"
	"os"
	gopath "path"
	"path/filepath"
	"strings"

	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/lib/entity"
	"github.com/surlykke/RefudeServices/lib/icon"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/response"
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
	entity.Base
	Name        string
	Type        string
	Permissions string
	Mimetype    string
	OsPath      string
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

func MakeLinkFromPath(ospath string, name string) link.Link {
	var title = name
	var mimetype, _ = magicmime.TypeByFile(ospath)
	var icon = icon.Name(strings.ReplaceAll(mimetype, "/", "-"))
	return link.Link{Href: "/file" + gopath.Clean(ospath), Title: title, Icon: icon, Relation: relation.Related, Type: mediatype.File}
}

func makeFileFromInfo(osPath string, fileInfo os.FileInfo) *File {
	var fileType = getFileType(fileInfo.Mode())
	var mimetype, _ = magicmime.TypeByFile(osPath)
	var icon = icon.Name(strings.ReplaceAll(mimetype, "/", "-"))
	var f = File{
		Base:        *entity.MakeBase(fileInfo.Name(), osPath, icon, mediatype.File),
		Name:        fileInfo.Name(),
		Type:        fileType,
		Permissions: fileInfo.Mode().String(),
		Mimetype:    mimetype,
		OsPath:      osPath,
	}

	for _, app := range applications.GetHandlers(f.Mimetype) {
		f.AddAction(app.DesktopId, app.Title, app.Icon)
	}
	return &f
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

func (f *File) DoPost(action string) response.Response {
	if action == "" && len(f.Actions) > 0 {
		action = f.Actions[0].Id
	}
	if applications.OpenFile(action, f.OsPath) {
		return response.Accepted()
	} else {
		return response.NotFound()
	}
}
