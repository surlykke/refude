package file

import (
	"fmt"
	"net/http"
	"net/url"
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

	var escapedPath = url.QueryEscape(f.Path)

	return &respond.StandardFormat{
		Self:         "/file?path=" + escapedPath,
		Type:         "file",
		Title:        f.Path,
		Comment:      comment,
		IconName:     strings.ReplaceAll(f.Mimetype, "/", "-"),
		OnPost:       "Open",
		OtherActions: "/file/actions?path=" + escapedPath,
		Data:         f,
	}
}

func (f *File) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, f.ToStandardFormat())
	} else if r.Method == "POST" {
		if err := applications.OpenFile(f.Path, f.Mimetype); err != nil {
			respond.ServerError(w, err)
		} else {
			respond.Accepted(w)
		}
	} else {
		respond.NotAllowed(w)
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

func (f *File) collectActions() respond.StandardFormatList {
	var sfl = make(respond.StandardFormatList, 0, 100)
	var recommendedApps, otherApps = applications.GetAppsForMimetype(f.Mimetype)
	for _, app := range recommendedApps {
		sf := makeFileAction(f, app).ToStandardFormat()
		sf.Type = "recommended"
		sfl = append(sfl, sf)
	}
	for _, app := range otherApps {
		sf := makeFileAction(f, app).ToStandardFormat()
		sf.Type = "other"
		sfl = append(sfl, sf)
	}
	return sfl
}

func (f *File) action(applicationId string) *FileAction {
	// Hmm. Maybe check if the application takes arguments?
	if da := applications.GetApp(applicationId); da == nil {
		return nil
	} else {
		return &FileAction{
			file:        f,
			application: da,
		}
	}
}

type FileAction struct {
	file        *File
	application *applications.DesktopApplication
}

func makeFileAction(f *File, app *applications.DesktopApplication) *FileAction {
	return &FileAction{
		file:        f,
		application: app,
	}
}

func (fa *FileAction) ToStandardFormat() *respond.StandardFormat {
	return &respond.StandardFormat{
		Self:     fmt.Sprintf("/file/action/%s?path=%s", fa.application.Id, fa.file.Path),
		Type:     "FileAction",
		Title:    fa.application.Name,
		Comment:  fa.application.Comment,
		IconName: fa.application.IconName,
	}
}

func (fa *FileAction) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, fa.ToStandardFormat)
	} else if r.Method == "POST" {
		if err := fa.application.Run(fa.file.Path); err != nil {
			respond.ServerError(w, err)
		} else {
			respond.Accepted(w)
			go applications.SetDefaultApp(fa.file.Mimetype, fa.application.Id)
		}
	} else {
		respond.NotAllowed(w)
	}
}
