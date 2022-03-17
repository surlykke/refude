// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/windows/x11"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/window/" {
		resource.ServeList(w, r, GetAll())
	} else {
		resource.ServeResource(w, r, Get(r.URL.Path))
	}
}

func GetAll() []resource.Resource {
	requestProxyMutex.Lock()
	defer requestProxyMutex.Unlock()
	var stackIdList = x11.GetStack(requestProxy)
	var resources = make([]resource.Resource, 0, len(stackIdList))
	for i, stackId := range stackIdList {
		var res = makeWindow(requestProxy, stackId)
		res.Stacking = i
		resources = append(resources, res)
	}

	return resources
}

func Get(path string) resource.Resource {
	if strings.HasPrefix(path, "/window/") {
		if i, err := strconv.Atoi(path[8:]); err == nil {
			if i >= 0 && i <= 0xFFFFFFFF {
				var windowId = uint32(i)
				requestProxyMutex.Lock()
				defer requestProxyMutex.Unlock()
				var stackIdList = x11.GetStack(requestProxy)
				for i, stackId := range stackIdList {
					if stackId == windowId {
						var v = makeWindow(requestProxy, windowId)
						v.Stacking = i
						return v
					}
				}
			}
		}
	}
	return nil
}

func Search(term string) link.List {
	requestProxyMutex.Lock()
	defer requestProxyMutex.Unlock()
	var stackIdList = x11.GetStack(requestProxy)
	var ll = make(link.List, 0, len(stackIdList))
	for _, wId := range stackIdList {
		var name, _ = x11.GetName(requestProxy, wId)
		if name != "org.refude.panel" {
			if rnk := searchutils.Match(term, name); rnk > -1 {
				var state = x11.GetStates(requestProxy, wId)
				if state&(x11.SKIP_TASKBAR|x11.SKIP_PAGER|x11.ABOVE) == 0 {
					var iconName, _ = GetIconName(requestProxy, wId)
					ll = append(ll, link.MakeRanked(fmt.Sprintf("/window/%d", wId), name, iconName, "window", rnk))
				}
			}
		}
	}
	return ll
}

func GetPaths() []string {
	requestProxyMutex.Lock()
	defer requestProxyMutex.Unlock()
	var wIdList = x11.GetStack(requestProxy)
	var paths = make([]string, 0, len(wIdList))
	for _, wId := range wIdList {
		paths = append(paths, fmt.Sprintf("/window/%d", wId))
	}
	return paths
}


func showAndRaise(id uint32) {
	requestProxyMutex.Lock()
	defer requestProxyMutex.Unlock()
	x11.MapAndRaiseWindow(requestProxy, id)
}

// http requests are concurrent, so all access to x11 from handling an http request, happens through
// this
var requestProxy = x11.MakeProxy()

// - and uses this for synchronization
var requestProxyMutex sync.Mutex
