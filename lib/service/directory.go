package service

import (
	"net/http"
	"github.com/surlykke/RefudeServices/lib/resource"
)

type directory []string

func (d directory) GET(w http.ResponseWriter, r *http.Request) {
	resource.JsonGET(d, w)
}


