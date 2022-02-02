// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package watch

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/surlykke/RefudeServices/lib/respond"
)

type event struct {
	next *event
	path string
}

var curr = &event{}

var lock = &sync.Mutex{}
var cond = sync.NewCond(lock)

func publish(path string) {
	lock.Lock()
	curr.next = &event{path: path}
	curr = curr.next
	lock.Unlock()
	cond.Broadcast()
}

func current() *event {
	lock.Lock()
	defer lock.Unlock()
	return curr
}

func next(ev *event) *event {
	lock.Lock()
	defer lock.Unlock()
	for ev.next == nil {
		cond.Wait()
	}
	return ev.next
}

func DesktopSearchMayHaveChanged() {
	publish("/search/desktop")
}

func SomethingChanged(path string) {
	publish(path)
}

const chunkTemplate = "%x\r\n" + // chunk length in hex
	"data:%s\n" +
	"\n" +
	"\r\n"

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respond.NotAllowed(w)
	} else if r.URL.Path == "/watch" {
		var ev = current()

		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.(http.Flusher).Flush()

		if _, err := fmt.Fprintf(w, "data:%s\n\n", ""); err != nil {
			return
		}
		w.(http.Flusher).Flush()

		for {
			ev = next(ev)
			if _, err := fmt.Fprintf(w, "data:%s\n\n", ev.path); err != nil {
				return
			}
			w.(http.Flusher).Flush()
		}

	} else {
		respond.NotFound(w)
	}
}
