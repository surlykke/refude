package file

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if file, err := makeFile(r.URL.Path[5:]); err != nil {
		log.Info("Could not make file from", r.URL.Path, err)
		respond.ServerError(w, err)
	} else if file != nil {
		var res = resource.MakeResource("/file"+file.Path, file.Name, "", file.Icon, "file", file)
		res.ServeHTTP(w, r)
	} else {
		respond.NotFound(w)
	}
}

func Collect(term string) link.List {
	var result = make(link.List, 0, 100)
	result = append(result, collectFrom(xdg.Home, term)...)
	result = append(result, collectFrom(xdg.DownloadDir, term)...)
	// Maybe more..
	return result
}

func collectFrom(searchDirectory, term string) link.List {
	var result = make(link.List, 0, 100)
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
		return nil
	}
	defer dir.Close()

	if names, err := dir.Readdirnames(-1); err != nil {
		log.Warn("Error reading", searchDirectory, err)
	} else {
		// Can't use filepath.Glob as it is case sensitive
		for _, name := range names {
			if rnk := searchutils.Match(term, name); rnk > -1 {
				var path = searchDirectory + "/" + name
				var mimetype, _ = magicmime.TypeByFile(path)
				var icon = strings.ReplaceAll(mimetype, "/", "-")
				result = append(result, link.MakeRanked("/file/"+path, prefix+name, icon, "file", rnk))
			}
		}
	}
	return result
}
