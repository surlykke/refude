package file

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

var filePathPattern = regexp.MustCompile(`^/file$|^/file(/actions)$|^/file/action/([^/]+)$`)

func Handler(r *http.Request) http.Handler {
	if r.URL.Path == "/file" {
		if file := getFile(r); file != nil {
			return file
		}
	}

	return nil
}

var noPathError = fmt.Errorf("No path given")

func getFile(r *http.Request) *File {
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

// Can't use filepath.Glob as it is case sensitive

func DesktopSearch(term string, baserank int) []respond.Link {
	if len(term) < 3 {
		return []respond.Link{}
	} else {
		term = strings.ToLower(term)
		var result = make([]respond.Link, 0, 100)
		for searchDirectory := range searchDirectories {
			if dir, err := os.Open(searchDirectory); err != nil {
				fmt.Println("Error opening", searchDirectory, err)
			} else if names, err := dir.Readdirnames(-1); err != nil {
				fmt.Println("Error reading", searchDirectory, err)
			} else {
				for _, name := range names {
					if rank, ok := searchutils.Rank(name, term, baserank); ok {
						if file, err := makeFile(searchDirectory + "/" + name); err != nil {
							fmt.Println("Error making file:", err)
						} else if file != nil {
							result = append(result, file.GetRelatedLink(rank))
						}

					}
				}
			}
		}
		return result
	}
}

func Recent(term string, baserank int) []respond.Link {
	var paths = getRecentDownloads(30 * time.Second)
	var result = make([]respond.Link, 0, len(paths))
	for _, path := range paths {
		if file, err := makeFile(path); err != nil {
			fmt.Println("Error making file:", err)
		} else if file != nil {
			if rank, ok := searchutils.Rank(file.Path, term, baserank); ok {
				result = append(result, file.GetRelatedLink(rank))
			}
		}
	}
	return result
}

type recentDownload struct {
	path       string
	downloaded time.Time
}

var recentDownloads = make([]recentDownload, 10)
var recentDownloadsLock sync.Mutex

func getRecentDownloads(noOlderThan time.Duration) []string {
	recentDownloadsLock.Lock()
	defer recentDownloadsLock.Unlock()

	var paths = make([]string, 0, len(recentDownloads))
	for _, recentDownload := range recentDownloads {
		if recentDownload.downloaded.After(time.Now().Add(-30 * time.Second)) {
			paths = append(paths, recentDownload.path)
		}
	}

	return paths
}

func addRecentDownload(path string) {
	recentDownloadsLock.Lock()
	defer recentDownloadsLock.Unlock()

	var newDownloads = make([]recentDownload, 0, len(recentDownloads)+1)
	var alreadyThere = false

	for _, rd := range recentDownloads {
		if rd.path == path {
			newDownloads = append(newDownloads, recentDownload{rd.path, time.Now()})
			alreadyThere = true
		} else if rd.downloaded.After(time.Now().Add(-time.Second * 30)) {
			newDownloads = append(newDownloads, rd)
		}
	}

	if !alreadyThere {
		newDownloads = append(newDownloads, recentDownload{path, time.Now()})
	}

	recentDownloads = newDownloads
}
