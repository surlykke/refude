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
)

// NotifierPath is reserved. Get requests to this path will
// be answered with a server-sent-event stream. Attemts to map
// a resource to NotifierPath will panic
const NotifierPath = "/notify"

// PingPath is reserved. Get request to this path will be answered with a http 200 ok
const PingPath = "/ping"


var	resources  map[string]Resource = make(map[string]Resource)
var	notifier   Notifier = MakeNotifier()
var mutex      sync.Mutex


type Resource interface {
	Data(r *http.Request) (int, string, []byte)
}


func Map(path string, res Resource) {
	mutex.Lock()
	defer mutex.Unlock()

	resources[path] = res
	notifier.Notify("resource-added", path)
}

func Remap(path string, res Resource) {
	mutex.Lock()
	defer mutex.Unlock()

	if _,ok := resources[path]; ok {
		resources[path] = res
		notifier.Notify("resource-updated", path)
	}
}

func Unmap(path string) {
	mutex.Lock()
	defer mutex.Unlock()

	if _,ok := resources[path]; ok {
		delete(resources, path)
		notifier.Notify("resource-removed", path)
	}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request for ", r.URL.Path)
	if r.URL.Path == NotifierPath {
		notifier.ServeHTTP(w, r)
	} else if r.URL.Path == PingPath {
		w.WriteHeader(http.StatusOK)
	} else {
		mutex.Lock()
		defer mutex.Unlock()

		if res,ok := resources[r.URL.Path]; ok {
			statusCode, contentType, bytes := res.Data(r)
			if statusCode != http.StatusOK {
				w.WriteHeader(statusCode)
			}
			if contentType != "" {
				w.Header().Set("Content-Type", contentType)
			}
			if bytes != nil {
				w.Write(bytes)
			}
		} else {
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



