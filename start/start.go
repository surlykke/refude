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
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/resource"
)

var lastUpdated = atomic.Pointer[time.Time]{}

type StartResource struct {
	resource.ResourceData
	searchTerm string
}

var startResource StartResource

func Run() {
	startResource = StartResource{ResourceData: *resource.MakeBase("/start", "Refude desktop", "", "", "start")}
	startResource.SetSearchHref("/search")
	repo.Put(&startResource)
}

func (s *StartResource) Search(term string) resource.LinkList {
	var result = make(resource.LinkList, 0, 100)
	getLinks(&result, "/notification/")
	getLinks(&result, "/window/")
	getLinks(&result, "/tab/")

	if len(term) > 0 {
		getLinks(&result, "/application/")
	}

	if len(term) > 2 {
		getLinks(&result, "/device/")
		result = append(result, file.SearchDesktop(term)...)
	}

	return result
}

func getLinks(collector *resource.LinkList, prefix string) {
	for _, res := range repo.GetListUntyped(prefix) {
		if !res.OmitFromSearch() {
			*collector = append(*collector, resource.LinkTo(res))
		}
	}
}
