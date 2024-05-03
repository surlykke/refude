// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package start

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/browsertabs"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/resourcerepo"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/wayland"
)

var lastUpdated = atomic.Pointer[time.Time]{}

type StartResource struct {
	resource.ResourceData
	searchTerm string
}

func (s *StartResource) Search(term string) []resource.Resource {
	term = strings.ToLower(term)
	var rrList = make(resource.RRList, 0, 200)

	wayland.Search(&rrList, term)
	browsertabs.Search(&rrList, term)
	notifications.Search(&rrList, term)
	applications.Search(&rrList, term)
	power.Search(&rrList, term)
	var t1 = time.Now()
	file.SearchDesktop(term, &rrList)
	var t2 = time.Now()
	fmt.Println("Search desktop took",  t2.Sub(t1))
	return rrList.GetResourcesSorted()
} 

func Run() {
	var start = &StartResource{ResourceData: *resource.MakeBase("/start", "Refude desktop", "", "", "start")}
	start.AddLink("/search", "", "", relation.Search)
	resourcerepo.Put(start)
}


