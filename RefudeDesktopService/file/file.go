package file

import (
	"path/filepath"
	"strings"

	"github.com/surlykke/RefudeServices/lib/respond"
)

type File struct {
	Path       string
	Mimetype   string
	DefaultApp string
}

func MakeFile(path, mimetype, defaultApp string) *File {
	return &File{
		Path:       path,
		Mimetype:   mimetype,
		DefaultApp: defaultApp,
	}
}

func (f *File) ToStandardFormat() *respond.StandardFormat {
	return &respond.StandardFormat{
		Self:         "/file?path=" + f.Path,
		Type:         "file",
		Title:        filepath.Base(f.Path),
		IconName:     strings.ReplaceAll(f.Mimetype, "/", "-"),
		OnPost:       "Launch with " + f.DefaultApp,
		OtherActions: "/file/actionsfor?path=" + f.Path,
		Data:         f,
	}
}
