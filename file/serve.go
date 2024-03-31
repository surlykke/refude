// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
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

type FileRepoType struct {}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/file/") {
		respond.NotFound(w)
	} else if file, err := makeFile(r.URL.Path[6:]); err != nil {
		respond.ServerError(w, err)
	} else if file == nil {
		respond.NotFound(w)
	} else {
		resource.ServeSingleResource(w, r, file,)
	}
}

func (fr FileRepoType) GetResource(filePath string) resource.Resource {
	if file, err := makeFile(filePath); err != nil {
		log.Warn("Could not make file from", filePath, err)
		return nil
	} else if file == nil {
		return nil
	} else {
		return file
	}
}

func (fr FileRepoType) Search(term string, threshold int) link.List{
	if len(term) < threshold {
		return link.List{}
	} else {
		return Search(xdg.Home, "~", term)
	}
}

var FileRepo FileRepoType
