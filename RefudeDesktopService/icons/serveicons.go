// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/lib/requests"

	"github.com/surlykke/RefudeServices/lib/resource"
)

var resources = func() *resource.GenericResourceCollection {
	var coll = resource.MakeGenericResourceCollection()
	coll.AddCollectionResource("/icons", "/icon/")
	coll.AddCollectionResource("/iconthemes", "/icontheme/")
	return coll
}()

/*
/icons
/icon/<name>
/icon/<name>/img
/icon/themeid/<name>
/icon/themeid/<name>/img

/iconthemes
/icontheme/<themeid>
*/

func Serve(w http.ResponseWriter, r *http.Request) bool {
	if strings.HasSuffix(r.URL.Path, "/img") {
		if res := resources.Get(r.URL.Path[0 : len(r.URL.Path)-4]); res != nil {
			var icon = res.(*Icon) // Will always succeed
			var size = uint32(32)
			if len(r.URL.Query()["size"]) > 0 {
				var ok bool
				if size, ok = readUint32(r.URL.Query()["size"][0]); !ok {
					requests.ReportUnprocessableEntity(w, fmt.Errorf("Invalid size"))
					return true
				}
			}
			fmt.Println(size, icon)
			var img = findImg(icon, size)
			// TODO Check Accept header
			http.ServeFile(w, r, img.Path)
			return true
		} else {
			return false
		}
	} else {
		return resource.ServeHttp(resources, w, r)
	}
}

func findImg(icon *Icon, size uint32) IconImage {
	var shortestDistanceSoFar = uint32(math.MaxUint32)
	var candidate IconImage

	for _, img := range icon.Images {
		var distance uint32
		if img.MinSize > size {
			distance = img.MinSize - size
		} else if img.MaxSize < size {
			distance = size - img.MaxSize
		} else {
			distance = 0
		}

		if distance < shortestDistanceSoFar {
			shortestDistanceSoFar = distance
			candidate = img
		}
		if distance == 0 {
			break
		}

	}

	return candidate
}
