// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"sync"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/windows/x11"
)

func Search(term string) link.List {
	return Windows.ExtractLinks(func(xWin XWin) int {
		proxyMutex.Lock()
		var name, _ = x11.GetName(synchronizedProxy, uint32(xWin))
		var state = x11.GetStates(synchronizedProxy, uint32(xWin))
		proxyMutex.Unlock()
		if name != "org.refude.panel" && state&(x11.SKIP_TASKBAR|x11.SKIP_PAGER|x11.ABOVE) == 0 {
			return searchutils.Match(term, name)
		} else {
			return -1
		}
	})
}


func showAndRaise(id uint32) {
	proxyMutex.Lock()
	defer proxyMutex.Unlock()
	x11.MapAndRaiseWindow(synchronizedProxy, id)
}

// http requests are concurrent, so all access to x11 from handling an http request, happens through
// this
var synchronizedProxy = x11.MakeProxy()

// - and uses this for synchronization
var proxyMutex sync.Mutex
