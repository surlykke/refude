// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
)

type Icon struct {
	Name   string
	Theme  string
	Images []IconImage
}

type IconImage struct {
	Context string
	MinSize uint32
	MaxSize uint32
	Path    string
}

type IconResource struct {
	resource.Links
}

func (ir *IconResource) GET(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else if name := requests.GetSingleQueryParameter(r, "name", ""); name == "" {
		requests.ReportUnprocessableEntity(w, errors.New("no name given"))
	} else {
		var themeId = requests.GetSingleQueryParameter(r, "theme", "hicolor")
		var size = uint64(32)
		var err error
		if len(r.URL.Query()["size"]) > 0 {
			size, err = strconv.ParseUint(r.URL.Query()["size"][0], 10, 32)
			if err != nil {
				requests.ReportUnprocessableEntity(w, errors.New("Invalid size given:"+r.URL.Query()["size"][0]))
			}
		}

		if image, ok := findImage(themeId, name, uint32(size)); !ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			http.ServeFile(w, r, image.Path)
		}
	}
}
