/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package service

import (
	"net/http"
	"fmt"
	"sync"
	"github.com/surlykke/RefudeServices/xdg"
	"net"
	"context"
	"syscall"
	"github.com/surlykke/RefudeServices/notify"
	"reflect"
)

// NotifierPath is reserved. Get requests to this path will
// be answered with a server-sent-event stream. Attemts to map
// a resource to NotifierPath will panic
const NotifierPath = "/notify"

// PingPath is reserved. Get request to this path will be answered with a http 200 ok
// Attempts to map to PingPath will panic
const PingPath = "/ping"

var	resources  map[string]http.Handler = make(map[string]http.Handler)
var mutex      sync.Mutex


func Map(path string, res http.Handler) {
	var eventType string

	mutex.Lock()
	if oldRes, ok := resources[path]; ok {
		if !reflect.DeepEqual(res, oldRes) {
			eventType = "resource-updated"
			resources[path] = res
		}
	} else {
		eventType = "resource-added"
		resources[path] = res
	}
	mutex.Unlock()

	if eventType != "" {
		notify.Notify(eventType, path[1:])
	}
}

func Unmap(path string) {
	var found bool

	mutex.Lock()
	if _,ok := resources[path]; ok {
		found = true
		delete(resources, path)
	}
	mutex.Unlock()

	if (found) {
		notify.Notify("resource-removed", path[1:])
	}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request for '", r.URL.Path, "'")
	if r.URL.Path == NotifierPath {
		notify.ServeHTTP(w, r)
	} else if r.URL.Path == PingPath {
		if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	} else {
		mutex.Lock()
		handlerCopy, ok := resources[r.URL.Path]
		mutex.Unlock()
		if ok {
			handlerCopy.ServeHTTP(w, r)
		} else {
			fmt.Println("Service doesn't have it..")
			w.WriteHeader(http.StatusNotFound)
		}
	}
}


func seemsToBeRunning(socketPath string) bool {
	client := http.Client{
		Transport: &http.Transport{ DialContext: func(ctx context.Context, _, _ string) (net.Conn, error){
				return net.Dial("unix", socketPath)
			},
		},
	}

	if response, err := client.Get("http://localhost/ping"); err == nil {
		response.Body.Close()
		return true
	} else {
		return false
	}
}

func makeListener(socketName string) (*net.UnixListener, bool) {
	socketPath := xdg.RuntimeDir() + "/" + socketName

	if seemsToBeRunning(socketPath) {
		fmt.Println("Application seems to be running. Let's leave it at that")
		return nil, false
	}

	syscall.Unlink(socketPath)

	if listener,err := net.ListenUnix("unix", &net.UnixAddr{Name: socketPath, Net: "unix"}); err != nil {
		fmt.Println(err)
		return nil, false
	} else {
		return listener, true
	}
}

func Serve(socketName string) {
	if listener, ok := makeListener(socketName); ok {
		http.Serve(listener, http.HandlerFunc(ServeHTTP))
	}
}

func ServeWith(socketName string, handler http.Handler) {
	if listener, ok := makeListener(socketName); ok {
		http.Serve(listener, handler)
	}
}



