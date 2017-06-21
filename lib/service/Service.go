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
	"github.com/surlykke/RefudeServices/lib/utils"
	"strings"
	"github.com/surlykke/RefudeServices/lib/resource"
	"log"
)


var	resources  = make(map[string]*resource.Resource)
var mutex      sync.Mutex


func OK200(this *resource.Resource, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func init() {
	Map("/ping", &resource.Resource{GET: OK200})
	Map("/notify", &resource.Resource{GET: notify.GET})
}


func splitInBaseNameAndName(path string) (string, string) {
	baseNameLen := strings.LastIndex(path[0:len(path) - 1], "/") + 1
	return path[0:baseNameLen], path[baseNameLen:]
}

// Caller must ensure path is proper and that
// mutex is taken
func _map(path string, res *resource.Resource) {
	baseName, name := splitInBaseNameAndName(path)
	if old, found := resources[path]; found {
		if !res.Equal(old) {
			resources[path] = res
			notify.Notify("resource-updated", path[1:])
		}
	} else {
		if len(baseName) > 0 {
			elements := []string{}
			if parent,ok := resources[baseName]; ok {
				elements = parent.Data.([]string)
			}
			elements = append(elements, name)

			_map(baseName, resource.JsonResource(elements, nil))
		}

		resources[path] = res
		notify.Notify("resource-added", path[1:])
	}
}

func unmap(path string) {
	baseName, name := splitInBaseNameAndName(path)
	delete(resources, path)
	notify.Notify("resource-removed", path[1:])
	if len(baseName) > 0 {
		parent := resources[baseName]
		elements := utils.Remove(parent.Data.([]string), name)
		resources[baseName] = resource.JsonResource(elements, nil)
		notify.Notify("resource-updated", baseName[1:])
	}
}

func checkPath(path string) {
	if strings.HasSuffix(path, "/") || !strings.HasPrefix(path, "/") {
		panic("Illegal path " + path + ". A path must begin with- and not end in '/'")
	}
}

func CreateEmptyDir(path string) bool {
	checkPath(path)
	mutex.Lock();
	defer mutex.Unlock();
	if _, ok := resources[path]; ok {
		return false
	} else {
		_map(path+"/", resource.JsonResource([]string{}, nil))
		return true
	}
}

func Map(path string, res *resource.Resource) {
	checkPath(path)
	mutex.Lock()
	defer mutex.Unlock()
	_map(path, res)
}

func Unmap(path string) *resource.Resource {
	checkPath(path)
	mutex.Lock()
	defer mutex.Unlock()
	if res,ok := resources[path]; ok {
		unmap(path)
		return res
	} else {
		return nil
	}
}

func UnMapIfMatch(path string, eTag string) *resource.Resource {
	checkPath(path)
	mutex.Lock()
	defer mutex.Unlock()
	if res,ok := resources[path]; ok {
		if res.ETag == eTag {
			unmap(path)
			return res
		}
	}

	return nil
}

func Has(path string) bool {
	checkPath(path)
	mutex.Lock()
	defer mutex.Unlock()
	_, has := resources[path]
	return has
}

func get(path string) (*resource.Resource, bool) {
	mutex.Lock()
	defer mutex.Unlock()
	res,ok := resources[path]
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
	socketPath := xdg.RuntimeDir + "/" + socketName

	if seemsToBeRunning(socketPath) {
		log.Fatal("Application seems to be running. Let's leave it at that")
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



