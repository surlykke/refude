// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package start

import (
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

var lastUpdated = atomic.Pointer[time.Time]{}

type StartResource struct {
	resource.ResourceData
	searchTerm string
}

var startResource StartResource

func Run() {
	startResource = StartResource{ResourceData: *resource.MakeBase("/start", "Refude desktop", "", "", mediatype.Start)}
	startResource.SetSearchHref("/search")
	repo.Put(&startResource)
}

func (s *StartResource) Search(term string) resource.LinkList {
	var result = make(resource.LinkList, 0, 100)
	if strings.Index(term, "/") > -1 {
		var pathBits = strings.Split(term, "/")
		pathBits, term = pathBits[:len(pathBits)-1], pathBits[len(pathBits)-1]
		var dirs = file.CollectDirs([]string{xdg.Home, xdg.ConfigHome, xdg.DownloadDir, xdg.DocumentsDir, xdg.MusicDir, xdg.VideosDir}, pathBits)
		for _, dir := range dirs {
			file.Collect(&result, dir)
		}
	} else {
		getLinks(&result, "/notification/")
		getLinks(&result, "/window/")
		getLinks(&result, "/tab/")

		if len(term) > 0 {
			getStartLinks(&result)
			getLinks(&result, "/application/")
		}

		if len(term) > 2 {
			getLinks(&result, "/device/")
			result = append(result, file.MakeLinkFromPath(xdg.Home, "Home"))
			for _, dir := range []string{xdg.Home, xdg.ConfigHome, xdg.DownloadDir, xdg.DocumentsDir, xdg.MusicDir, xdg.VideosDir} {
				file.Collect(&result, dir)
			}

		}
	}

	return result.FilterAndSort(term)
}

func getLinks(collector *resource.LinkList, prefix string) {
	for _, res := range repo.GetListUntyped(prefix) {
		if !res.OmitFromSearch() {
			*collector = append(*collector, resource.LinkTo(res))
		}
	}
}

func (s StartResource) DoPost(w http.ResponseWriter, r *http.Request) {
	var action = requests.GetSingleQueryParameter(r, "action", "-")
	if exec, ok := getExec(action); ok {
		if err := xdg.RunCmd(exec...); err != nil {
			log.Warn(err)
			respond.ServerError(w, err)
		} else {
			respond.Accepted(w)
		}
	} else {
		respond.NotFound(w)
	}
}
