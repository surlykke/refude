package file

import (
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if filePath, err := url.PathUnescape(r.URL.Path[6:]); err != nil {
		log.Info("Could not extract path from", r.URL.Path, err)
	} else if file, err := makeFile(filePath); err != nil {
		log.Info("Could not make file from", filePath, err)
	} else {
		var res = resource.MakeResource("/file/"+url.PathEscape(file.Path), file.Name, "", file.Icon, "file", file)
		res.Links = res.Links.Filter(requests.Term(r))
		res.ServeHTTP(w, r)
		return
	}
	respond.NotFound(w)
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
