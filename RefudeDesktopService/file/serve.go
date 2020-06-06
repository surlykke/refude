package file

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/applications"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("file.ServeHTTP, query:", r.URL.Query())
	if r.URL.Path == "/file" {
		if file, err := getFile(r); err != nil {
			respond.ServerError(w, err)
		} else if file == nil {
			respond.NotFound(w)
		} else if r.Method == "POST" {
			if err := applications.OpenFile(file.Path, file.Mimetype); err != nil {
				respond.ServerError(w, err)
			} else {
				respond.Accepted(w)
			}
		} else {
			respond.AsJson(w, r, file.ToStandardFormat())
		}
	} else if r.URL.Path == "/file/otheractions" {
		if file, err := getFile(r); err != nil {
			respond.ServerError(w, err)
		} else if file == nil {
			respond.NotFound(w)
		} else {
			respond.AsJson(w, r, otherActions(file, searchutils.Term(r)))
		}
	} else if r.URL.Path == "/file/action" {
		if file, err := getFile(r); err != nil {
			respond.ServerError(w, err)
		} else if file == nil {
			respond.NotFound(w)
		} else if app := getApp(r); app == nil {
			respond.NotFound(w)
		} else {
			if r.Method == "POST" {
				if err := app.Run(file.Path); err != nil {
					respond.ServerError(w, err)
				} else {
					respond.Accepted(w)
					go applications.SetDefaultApp(file.Mimetype, app.Id)
				}
			} else {
				respond.AsJson(w, r, buildActionSF(file, app, "action"))
			}
		}
	} else {
		respond.NotFound(w)
	}
}

var noPathError = fmt.Errorf("No path given")

func getFile(r *http.Request) (*File, error) {
	if path := getAdjustedPath(r); path == "" {
		return nil, noPathError
	} else if file, err := makeFile(path); err != nil {
		return nil, err
	} else if file == nil {
		return nil, nil
	} else {
		return file, nil
	}
}

func getApp(r *http.Request) *applications.DesktopApplication {
	return applications.GetApp(requests.GetSingleQueryParameter(r, "app", ""))
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
			DefaultApp: "TODO",
		}, nil
	}
}

func getAdjustedPath(r *http.Request) string {
	if path := requests.GetSingleQueryParameter(r, "path", ""); path == "" {
		return ""
	} else if path[0] != '/' {
		return xdg.Home + "/" + path
	} else {
		fmt.Println("file.getAdjustedPath returning:", path)
		return path
	}

}

func otherActions(f *File, term string) respond.StandardFormatList {
	var sfl = make(respond.StandardFormatList, 0, 100)
	var recommendedApps, otherApps = applications.GetAppsForMimetype(f.Mimetype)
	for _, app := range recommendedApps {
		var sf = buildActionSF(f, app, "recommended")
		if sf.Rank = searchutils.SimpleRank(sf.Title, sf.Comment, term); sf.Rank > -1 {
			sfl = append(sfl, sf)
		}
	}
	sfl.SortByRank()
	var numRecommended = len(sfl)
	for _, app := range otherApps {
		var sf = buildActionSF(f, app, "other")
		if sf.Rank = searchutils.SimpleRank(sf.Title, sf.Comment, term); sf.Rank > -1 {
			sfl = append(sfl, sf)
		}
	}

	sfl[numRecommended:].SortByRank()

	return sfl
}

func buildActionSF(f *File, app *applications.DesktopApplication, Type string) *respond.StandardFormat {
	return &respond.StandardFormat{
		Self:     appActionPath(f, app),
		Type:     Type,
		Title:    app.Name,
		Comment:  app.Comment,
		IconName: app.IconName,
	}
}

var searchDirectories []string

func init() {
	var added = make(map[string]bool)
	var dirsToAdd = []string{
		xdg.Home,
		xdg.DesktopDir,
		xdg.DownloadDir,
		xdg.TemplatesDir,
		xdg.PublicshareDir,
		xdg.DocumentsDir,
		xdg.MusicDir,
		xdg.PicturesDir,
		xdg.VideosDir,
	}
	for _, d := range dirsToAdd {
		if !added[d] {
			searchDirectories = append(searchDirectories, d)
			added[d] = true
		}
	}
}

func DesktopSearch(term string) respond.StandardFormatList {
	var result = make(respond.StandardFormatList, 0, 100)
	for _, searchDirectory := range searchDirectories {
		if dir, err := os.Open(searchDirectory); err != nil {
			fmt.Println("Error opening", searchDirectory, err)
		} else if names, err := dir.Readdirnames(-1); err != nil {
			fmt.Println("Error reading", searchDirectory, err)
		} else {
			for _, name := range names {
				if rank := searchutils.SimpleRank(name, "", term); rank > -1 {
					if file, err := makeFile(searchDirectory + "/" + name); err != nil {
						fmt.Println("Error making file:", err)
					} else if file != nil {
						result = append(result, file.ToStandardFormat().Ranked(rank))
					}
				}
			}
		}
	}

	return result
}

func fileSelf(file *File) string {
	return "/file?path=" + url.QueryEscape(file.Path)
}

func otherActionsPath(file *File) string {
	return "/file/otheractions?path=" + url.QueryEscape(file.Path)
}

func appActionPath(file *File, app *applications.DesktopApplication) string {
	return "/file/action?path=" + url.QueryEscape(file.Path) + "&app=" + url.QueryEscape(app.Id)
}
