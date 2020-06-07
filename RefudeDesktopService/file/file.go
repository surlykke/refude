package file

import (
	"os"
	"strings"

	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/applications"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type File struct {
	Path       string
	Mimetype   string
	DefaultApp string
}

func (f *File) ToStandardFormat() *respond.StandardFormat {
	var comment string

	if f.DefaultApp != "" {
		comment = "Open with " + f.DefaultApp
	} else {
		comment = "Open"
	}

	return &respond.StandardFormat{
		Self:         fileSelf(f),
		Type:         "file",
		Title:        f.Path,
		Comment:      comment,
		IconName:     strings.ReplaceAll(f.Mimetype, "/", "-"),
		OnPost:       "Open",
		OtherActions: otherActionsPath(f),
		Data:         f,
	}
}

func makeFile(path string) (*File, error) {
	if !strings.HasPrefix(path, "/") {
		path = xdg.Home + "/" + path
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else {
		var mimetype, _ = magicmime.TypeByFile(path)
		return &File{
			Path:       path,
			Mimetype:   mimetype,
			DefaultApp: applications.GetDefaultAppName(mimetype),
		}, nil
	}
}
