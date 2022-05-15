// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package file

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if f := Get(r.URL.Path[5:]); f != nil {
		var self = "/file/" + f.Id()
		resource.ServeResource[string](w, r, self, f)
	} else {
		respond.NotFound(w)
	}
}

func Get(filePath string) *File {
	if file, err := makeFile(filePath); err != nil {
		log.Warn("Could not make file from", filePath, err)
		return nil
	} else if file == nil {
		return nil
	} else {
		return file
	}
}
