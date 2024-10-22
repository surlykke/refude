// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package file

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/path"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if f := GetResource(path.Of(r.URL.Path)); f == nil {
		respond.NotFound(w)
	} else {
		resource.ServeSingleResource(w, r, f)
	}
}

func GetResource(resPath path.Path) *File {
	var pathS = string(resPath)
	fmt.Println("file.GetResource, path: '" + resPath + "'")
	if !strings.HasPrefix(pathS, "/file/") {
		log.Warn("Unexpeded path:", resPath)
		return nil
	} else if file, err := makeFileFromPath(pathS[5:]); err != nil {
		log.Warn("Could not make file from", pathS[5:], err)
		return nil
	} else if file == nil {
		fmt.Println(".. not found")
		return nil
	} else {
		return file
	}
}
