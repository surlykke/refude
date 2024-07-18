// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package start

import (
	"sync/atomic"
	"time"

	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

var lastUpdated = atomic.Pointer[time.Time]{}

type StartResource struct {
	resource.ResourceData
	searchTerm string
}

var startResource StartResource

func Run() {
	startResource = StartResource{ResourceData: *resource.MakeBase("/start", "Refude desktop", "", "", "start")}
	startResource.AddLink("/search", "", "", relation.Search)
	repo.Put(&startResource)
}

func (s *StartResource) Search(term string) []resource.Resource {
	var result = make([]resource.Resource, 0, 100)
	result = append(result, searchList(repo.GetListUntyped("/notification/"), term)...)
	result = append(result, searchList(repo.GetListUntyped("/window/"), term)...)
	result = append(result, searchList(repo.GetListUntyped("/tab/"), term)...)
	if len(term) > 0 {
		result = append(result, searchList(repo.GetListUntyped("/application/"), term)...)
	}

	if len(term) > 2 {
		result = append(result, searchList(repo.GetListUntyped("/device/"), term)...)
		result = append(result, file.SearchDesktop(term).GetResources()...)
	}

	return result
}

func searchList(list []resource.Resource, term string) []resource.Resource {
	var rrList = make(resource.RRList, 0, len(list))
	for _, res := range list {
		if res.OmitFromSearch() {
			continue
		}
		if rnk := searchutils.Match(term, res.GetTitle(), res.GetKeywords()...); rnk >= 0 {
			rrList = append(rrList, resource.RankedResource{Rank: rnk, Res: res})
		}
	}
	var resources = rrList.GetResources()
	return resources
}
