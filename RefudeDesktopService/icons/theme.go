package icons

import (
	"encoding/json"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/resource"
)

type Theme struct {
	resource.Links
	Id       string
	Name     string
	Comment  string
	Inherits []string
	Dirs     map[string]IconDir
}

type IconDir struct {
	Path    string
	MinSize uint32
	MaxSize uint32
	Context string
}

func (t *Theme) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var json, err = json.Marshal(t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Write(json)
	}
}
