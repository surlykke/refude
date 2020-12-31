package windows

import "C"

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/respond"
)

type Monitor struct {
	respond.Links `json:"_links"`
	X, Y          int32
	W, H          uint32
	Wmm, Hmm      uint32
	Name          string
}

func (m *Monitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, m)
	} else {
		respond.NotAllowed(w)
	}
}
