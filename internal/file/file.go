// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package file

import (
	"io/fs"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/surlykke/refude/internal/applications"
	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/pkg/bind"
)

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

func MimeType(ospath string) string {
	var ext = filepath.Ext(ospath)
	if ext != "" {
		return mime.TypeByExtension(ext)
	} else {
		return ""
	}
}

func makeFileFromInfo(osPath string, fileInfo os.FileInfo) *File {
	var fileType = getFileType(fileInfo.Mode())
	var mimetype string
	if "Directory" == fileType {
		mimetype = "inode/directory"
	} else {
		mimetype = MimeType(osPath)
	}
	var icon = strings.ReplaceAll(mimetype, "/", "-")
	var f = File{
		Base:        *entity.MakeBase(fileInfo.Name(), osPath, icon, "File"),
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
		log.Print("Could not open", dir, err)
		return nil
	} else if entries, err := file.ReadDir(-1); err != nil {
		log.Print("Could not read", dir, err)
		return nil
	} else {
		return entries
	}
}

func (f *File) DoPost(action string) bind.Response {
	if action == "" && len(f.Meta.Actions) > 0 {
		action = f.Meta.Actions[0].Id
	}
	if applications.OpenFile(action, f.OsPath) {
		return bind.Accepted()
	} else {
		return bind.NotFound()
	}
}
