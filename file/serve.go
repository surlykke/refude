package file

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

var filePathPattern = regexp.MustCompile(`^/file$|^/file(/actions)$|^/file/action/([^/]+)$`)

var noPathError = fmt.Errorf("No path given")

func GetResource(r *http.Request) (resource.Resource, bool) {
	if filePath, err := url.PathUnescape(r.URL.Path[6:]); err != nil {
		log.Info("Could not extract path from", r.URL.Path, err)
	} else if file, err := makeFile(filePath); err != nil {
		log.Info("Could not make file from", filePath, err)
	} else {
		var res = resource.Make("/file/"+url.PathEscape(file.Path), file.Name, "", file.Icon, "file", file)
		res.Links = res.Links.Filter(requests.Term(r))
		return res, true
	}
	return resource.Resource{}, false
}

func Collect(term string, sink chan link.Link) {
	collectFrom(xdg.Home, term, sink)
	collectFrom(xdg.DownloadDir, term, sink)
	// Maybe more..
}

func collectFrom(searchDirectory, term string, sink chan link.Link) {
	var prefix string
	if searchDirectory == xdg.Home {
		prefix = "~/"
	} else {
		prefix = path.Base(searchDirectory) + "/"
	}

	var dir *os.File
	var err error
	if dir, err = os.Open(searchDirectory); err != nil {
		log.Warn("Error opening", searchDirectory, err)
		return
	}

	if names, err := dir.Readdirnames(-1); err != nil {
		log.Warn("Error reading", searchDirectory, err)
	} else {
		// Can't use filepath.Glob as it is case sensitive
		for _, name := range names {
			if rnk := searchutils.Match(term, name); rnk > -1 {
				var path = searchDirectory + "/" + name
				var mimetype, _ = magicmime.TypeByFile(path)
				var icon = strings.ReplaceAll(mimetype, "/", "-")
				sink <- link.MakeRanked("/file/"+url.PathEscape(path), prefix+name, icon, "file", rnk)
			}
		}
	}

	dir.Close()
}
