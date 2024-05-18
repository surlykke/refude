// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package start

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/resource"
)

var lastUpdated = atomic.Pointer[time.Time]{}

type StartResource struct {
	resource.ResourceData
	searchTerm string
}

var startResource = StartResource{ResourceData: *resource.MakeBase("/start", "Refude desktop", "", "", "start")}


func Run() {
	var startRequests = repo.MakeAndRegisterRequestChan()
	startResource.AddLink("/search", "", "", relation.Search)
	for req := range startRequests {
		if req.ReqType == repo.ByPath && req.Data == "/start" || req.ReqType == repo.ByPathPrefix && strings.HasPrefix("/start", req.Data) {
			req.Replies <- resource.RankedResource{Res: &startResource}
		}
		req.Wg.Done()
	}
}

func (s *StartResource) Search(term string) []resource.Resource {
	return repo.DoSearch(term)
} 



