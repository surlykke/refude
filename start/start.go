// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package start

import (
	"strings"

	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/resourcerepo"
)

type StartResource struct {
	resource.BaseResource
	searchTerm string
}

func (s *StartResource) Links(term string) link.List {
	return DoDesktopSearch(term)
}

func DoDesktopSearch(term string) link.List {
	var links = make(link.List, 0, 300)
	term = strings.ToLower(term)
	links = append(links, resourcerepo.Search(term)...)
	links = append(links, file.FileRepo.Search(term, 2)...)

	return links
}

func Run() {
	resourcerepo.Put(&StartResource{BaseResource: resource.BaseResource{Path: "/start", Title: "Start", Profile: "start"}})
}


