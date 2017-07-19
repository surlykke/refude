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
	"github.com/surlykke/RefudeServices/lib/utils"
	"strings"
	"github.com/surlykke/RefudeServices/lib/resource"
	"log"
	"reflect"
)


var	resources  = make(map[string]interface{})
var mutex      sync.Mutex

type PingResource struct {
}

func (pr* PingResource) GET(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func init() {
	Map("/ping", &PingResource{})

	Map("/notify", &NotifyResource{})
}

func panicIfPathNotOk(path string) {
	if !checkPath(path) {
		panic("Illegal path " + path + ". A path must begin with- and not end in '/'")
	}
}

func checkPath(path string) bool {
	return 	strings.HasPrefix(path, "/") && !strings.HasSuffix(path, "/")
}

func split(path string) (string, string) {
	dirPathLen := strings.LastIndex(path[0:len(path) - 1], "/") + 1
	return path[0:dirPathLen], path[dirPathLen:]
}

// "", "applications"

func addEntryToDir(dirPath string, entry string) {
	pathWithoutSlash := dirPath[0:len(dirPath) - 1]
	var dir directory
	if res, ok := resources[dirPath]; !ok {
		if _,ok := resources[pathWithoutSlash]; ok {
			log.Fatal("Not a directory", dirPath)
		}
		if dirPath != "/" {
			addEntryToDir(split(dirPath))
		}
		dir = directory{entry}
		ResourceAdded(dirPath)
	} else {
		dir = directory(append(utils.Copy(res.(directory)), entry))
		ResourceUpdated(dirPath)
	}

	resources[dirPath] = dir
}

func mapEntry(path string, res interface{}) {
	if _,ok := resources[path + "/"]; ok {
		log.Fatal("There's a directory there: ", path)
	} else if oldRes, ok := resources[path]; ok {
		if !reflect.DeepEqual(oldRes, res) {
			ResourceUpdated(path)
			resources[path] = res
		}
	} else {
		addEntryToDir(split(path))
		ResourceAdded(path)
		resources[path] = res
	}
}

func unmapEntry(path string) bool {
	if res,ok := resources[path]; ok {
		dirPath, entry := split(path)
		if _,ok := res.(directory); ok {
			log.Fatal("Unmapping directory path", path)
		}
		delete(resources, path)
		ResourceRemoved(path)
		dir := directory(utils.Remove(resources[dirPath].(directory), entry))
		resources[dirPath] = dir
		ResourceUpdated(dirPath)
		return true
	} else {
		return false
	}
}


func MkDir(path string) {
	panicIfPathNotOk(path)
	mutex.Lock()
	defer mutex.Unlock()
	mapEntry(path + "/", directory{})
}

func Map(path string, res interface{}) {
	panicIfPathNotOk(path)
	mutex.Lock()
	defer mutex.Unlock()
	mapEntry(path, res)
}

func Unmap(path string) bool {
	panicIfPathNotOk(path)
	mutex.Lock()
	defer mutex.Unlock()
	return unmapEntry(path)
}

func UnMapIfMatch(path string, eTag string) bool {
	panicIfPathNotOk(path)
	mutex.Lock()
	defer mutex.Unlock()
	if res,ok := resources[path]; ok {
		if etagHandler, ok := res.(resource.ETagHandler); ok {
			if etagHandler.ETag() == eTag {
				unmapEntry(path)
				return true
			}
		}
	}
	return false;
}

func Has(path string) bool {
	panicIfPathNotOk(path)
	mutex.Lock()
	defer mutex.Unlock()
	_, has := resources[path]
	return has
}

func Get(path string) (interface{}, bool) {
	panicIfPathNotOk(path)
	return get(path)
}

func get(path string) (interface{}, bool) {
	mutex.Lock()
	defer mutex.Unlock()
	i, ok := resources[path]
	return i,ok
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if res,ok := get(r.URL.Path); ok {
		resource.ServeHTTP(res, w, r)
	} else if res, ok = get(r.URL.Path + "/"); ok {
		resource.ServeHTTP(res, w, r)
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



