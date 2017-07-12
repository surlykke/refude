// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
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

func checkPath(path string) {
	if strings.HasSuffix(path, "/") || !strings.HasPrefix(path, "/") {
		panic("Illegal path " + path + ". A path must begin with- and not end in '/'")
	}
}

func splitInBaseNameAndName(path string) (string, string) {
	baseNameLen := strings.LastIndex(path[0:len(path) - 1], "/") + 1
	return path[0:baseNameLen], path[baseNameLen:]
}

func addEntryToDir(dirPath string, entry string) {
	var dir []string
	if res, ok := resources[dirPath]; !ok {
		dir = []string{entry}
		if dirPath != "/" {
			addEntryToDir(splitInBaseNameAndName(dirPath))
		}
		notify.ResourceAdded(dirPath)
	} else if dir, ok = res.Data.([]string); !ok {
		log.Fatal("Adding entry", entry, "to nondir", dirPath)
	} else {
		dir = append(dir, entry)
	}
	resources[dirPath] = resource.JsonResource(dir, nil)
	notify.ResourceUpdated(dirPath)
}

func mapEntry(path string, res *resource.Resource) {
	if oldRes,ok := resources[path]; ok {
		if _,ok = oldRes.Data.([]string); ok {
			log.Fatal("Directory exists", path)
		} else if !oldRes.Equal(res) {
			notify.ResourceUpdated(path)
		}
	} else {
		addEntryToDir(splitInBaseNameAndName(path))
		notify.ResourceAdded(path)
	}
	resources[path] = res
}


func unmapEntry(path string) bool {
	if res,ok := resources[path]; ok {
		dirPath, entry := splitInBaseNameAndName(path)
		if _,ok := res.Data.([]string); ok {
			log.Fatal("Unmapping directory path", path)
		}
		delete(resources, path)
		notify.ResourceRemoved(path)
		dir := utils.Remove(resources[dirPath].Data.([]string), entry)
		resources[dirPath] = resource.JsonResource(dir, nil)
		notify.ResourceUpdated(dirPath)
		return true
	} else {
		return false
	}
}


func MkDir(path string) {
	checkPath(path)
	mapEntry(path, resource.JsonResource([]string{}, nil))
}

func Map(path string, res *resource.Resource) {
	checkPath(path)
	mutex.Lock()
	defer mutex.Unlock()
	mapEntry(path, res)
}

func Unmap(path string) bool {
	checkPath(path)
	mutex.Lock()
	defer mutex.Unlock()
	return unmapEntry(path)
}

func UnMapIfMatch(path string, eTag string) bool {
	checkPath(path)
	mutex.Lock()
	defer mutex.Unlock()
	if res,ok := resources[path]; ok && res.ETag == eTag {
		unmapEntry(path)
		return true
	} else {
		return false;
	}
}

func Has(path string) bool {
	checkPath(path)
	mutex.Lock()
	defer mutex.Unlock()
	_, has := resources[path]
	return has
}

func Get(path string) (*resource.Resource, bool) {
	checkPath(path)
	mutex.Lock()
	defer mutex.Unlock()
	res, ok := resources[path]
	return res,ok
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handlerCopy, ok := resources[r.URL.Path]; ok {
		handlerCopy.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func seemsToBeRunning(socketPath string) bool {
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error){
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

	listener,err := net.ListenUnix("unix", &net.UnixAddr{
		Name: socketPath,
		Net: "unix",
	});

	if err != nil {
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



