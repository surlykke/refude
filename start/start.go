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

	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/resourcerepo"
)

var lastUpdated = atomic.Pointer[time.Time]{}

type StartResource struct {
	resource.BaseResource
	searchTerm string
}

func (s *StartResource) Search(term string) []resource.Resource {
	return DoDesktopSearch(term)
}

func (s *StartResource) UpdatedSince(t time.Time) bool {
	return t.Before(*lastUpdated.Load())
}

func DoDesktopSearch(term string) []resource.Resource {
	var resList = make([]resource.Resource, 0, 300)
	term = strings.ToLower(term)
	resList = append(resList, resourcerepo.Search(term)...)
	resList = append(resList, file.FileRepo.Search(term, 2)...)
	return resList
}

func Run() {
	var start = &StartResource{BaseResource: *resource.MakeBase("/start", "Refude desktop", "", "", "start")}
	start.AddLink("/search", "", "", relation.Search)
	resourcerepo.Put(start)
}


