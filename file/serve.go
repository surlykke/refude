package file

import (
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if file, err := makeFile(r.URL.Path[5:]); err != nil {
		log.Info("Could not make file from", r.URL.Path, err)
		respond.ServerError(w, err)
	} else if file != nil {
		var title = file.Path
		if strings.HasPrefix(title, xdg.Home) {
			title = "~" + title[len(xdg.Home):]
		}
		var res = resource.MakeResource("/file"+file.Path, title, file.Type, file.Icon, "file", file)
		res.ServeHTTP(w, r)
	} else {
		respond.NotFound(w)
	}
}

func Search(term string) link.List {
	if file, err := makeFile(xdg.Home); err != nil {
		return link.List{}
	} else{
		return file.GetLinks(term)
	}
}


