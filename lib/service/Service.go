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
	"github.com/surlykke/RefudeServices/lib/xdg"
	"net"
	"context"
	"syscall"
	"github.com/surlykke/RefudeServices/lib/notify"
	"reflect"
	"github.com/surlykke/RefudeServices/lib/stringlist"
	"strings"
	"log"
)

type PingResource struct{}

func (pr PingResource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func init() {
	Map("/ping", PingResource{})
	Map("/notify", notify.NotifierResource{})
}


var	resources  map[string]http.Handler = make(map[string]http.Handler)
var mutex      sync.Mutex


// Adds/owerwrites a resource. Adds/owerwrites containing directories as necessary
// Returns a list of modified resources (not created)
func addResource(path string, resource http.Handler) stringlist.StringList {
	if !strings.HasPrefix(path, "/") {
		log.Println("Attempt to map", path, "- paths must begin with '/'")
		return []string{}
	}

	if strings.HasSuffix(path, "/") {
		log.Println("Attempt to map", path, "- paths must not end in '/'")
		return []string{}
	}

	updated := make(stringlist.StringList, 0)

	mutex.Lock()
	defer mutex.Unlock()

	if oldRes,ok := resources[path]; ok && !reflect.DeepEqual(oldRes, resource) {
		updated = stringlist.PushBack(updated, path)
	}
	resources[path] = resource

	for len(path) > 0 {
		lastSlashPos := strings.LastIndex(path[0: len(path) - 1], "/")
		if lastSlashPos < 0 {
			break
		}
		directory, element := path[0:lastSlashPos + 1],path[lastSlashPos + 1:]
		if elements,ok := resources[directory].(stringlist.StringList); ok {
			if !elements.Has(element) {
				updated = stringlist.PushBack(updated, directory)
				resources[directory] = stringlist.PushBack(elements, element)
			}
			break
		} else {
			resources[directory] = stringlist.StringList{element}
			path = directory
		}
	}

	return updated
}

func removeResource(path string) (string, string) {
	if !strings.HasPrefix(path, "/") {
		log.Println("Attempt to unmap", path, "- paths must begin with '/'")
		return "",""
	}

	if strings.HasSuffix(path, "/") {
		log.Println("Attempt to unmap", path, "- paths must not end in '/'")
		return "",""
	}

	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := resources[path]; ok {
		delete(resources, path)
		lastSlashPos := strings.LastIndex(path, "/")
		directory, element := path[0: lastSlashPos + 1], path[lastSlashPos + 1:]
		elements,_ := resources[directory].(stringlist.StringList)
		elements = stringlist.Remove(elements, element)
		resources[directory] = elements
		// TOCONSIDER: Should we auto-unmap empty directories?

		return path, directory
	} else {
		return "", ""
	}
}

func Map(path string, res http.Handler) {
	updated := addResource(path, res)

	for _,path := range updated {
		notify.Notify("resource-updated", "." + path)
	}
}

func Unmap(path string) {
	removed, updated := removeResource(path)

	if removed != "" {
		notify.Notify("resource-removed", "." + removed)
	}

	if updated != "" {
		notify.Notify("resource-updated", "." + updated)
	}
}

func Has(path string) bool {
	_,has := get(path)
	return has
}

func get(path string) (http.Handler, bool) {
	mutex.Lock()
	defer mutex.Unlock()
	res, ok := resources[path]
	return res,ok
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handlerCopy, ok := get(r.URL.Path); ok {
		handlerCopy.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
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



