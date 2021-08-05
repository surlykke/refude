package file

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

var filePathPattern = regexp.MustCompile(`^/file$|^/file(/actions)$|^/file/action/([^/]+)$`)

var noPathError = fmt.Errorf("No path given")

func GetResource(pathElements []string) resource.Resource {
	if len(pathElements) == 1 {
		if filePath, err := url.PathUnescape(pathElements[0]); err != nil {
			log.Info("Could not extract path from", pathElements[0], err)
		} else if file, err := makeFile(filePath); err != nil {
			log.Info("Could not make file from", filePath, err)
		} else {
			return file
		}
	}
	return nil
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
					var icon = strings.ReplaceAll(mimetype, "/", "-")
					crawler("/file/"+url.PathEscape(path), path, icon)
				}
			}
		}
		dir.Close()
	}
}
