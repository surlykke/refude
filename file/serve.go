package file

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

var filePathPattern = regexp.MustCompile(`^/file$|^/file(/actions)$|^/file/action/([^/]+)$`)

var noPathError = fmt.Errorf("No path given")

func GetJsonResource(r *http.Request) respond.JsonResource {
	if path := getAdjustedPath(r); path == "" {
		return nil
	} else if file, err := makeFile(path); err != nil {
		return nil
	} else if file == nil {
		return nil
	} else {
		return file
	}
}

func getAdjustedPath(r *http.Request) string {
	if path := requests.GetSingleQueryParameter(r, "path", ""); path == "" {
		return ""
	} else if path[0] != '/' {
		return xdg.Home + "/" + path
	} else {
		return path
	}

}

var searchDirectories = make(map[string]bool, 9)

func init() {
	searchDirectories[xdg.Home] = true
	searchDirectories[xdg.DesktopDir] = true
	searchDirectories[xdg.DownloadDir] = true
	searchDirectories[xdg.TemplatesDir] = true
	searchDirectories[xdg.PublicshareDir] = true
	searchDirectories[xdg.DocumentsDir] = true
	searchDirectories[xdg.MusicDir] = true
	searchDirectories[xdg.PicturesDir] = true
	searchDirectories[xdg.VideosDir] = true
}

func Crawl(term string, forDisplay bool, crawler searchutils.Crawler) {
	var termRunes = []rune(term)
	for searchDirectory := range searchDirectories {
		var dir *os.File
		var err error
		if dir, err = os.Open(searchDirectory); err != nil {
			log.Warn("Error opening", searchDirectory, err)
			continue
		}

		if names, err := dir.Readdirnames(-1); err != nil {
			log.Warn("Error reading", searchDirectory, err)
		} else {
			// Can't use filepath.Glob as it is case sensitive
			for _, name := range names {
				if searchutils.FluffyIndex([]rune(strings.ToLower(name)), termRunes) > -1 {
					var path = searchDirectory + "/" + name
					var mimetype, _ = magicmime.TypeByFile(path)
					var resource = makeResource(path, mimetype)
					crawler(&resource, nil)
				}

			}
		}
		dir.Close()
	}
}
