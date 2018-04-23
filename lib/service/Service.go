// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package service

import (
	"net/http"
	"sync"
	"log"
	"github.com/surlykke/RefudeServices/lib/resource"
	"context"
	"net"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"syscall"
	"fmt"
)


var	resources  = make(map[string]interface{})
var mutex      sync.Mutex
var root       = MakeEmptyDir()

type PingResource struct {
}

func (pr* PingResource) GET(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}


func init() {
	Map("/ping", &PingResource{})
}

func MkDir(path string) {
	root.MkDir(Standardize(path))
}


func Map(path string, res interface{}) {
	mutex.Lock()
	defer mutex.Unlock()
	root.Map(Standardize(path), res)
}

func Unmap(path string) {
	mutex.Lock()
	defer mutex.Unlock()
	root.UnMap(Standardize(path))
}

func UnMapIfMatch(path string, eTag string) bool {
	sp := Standardize(path)
	mutex.Lock()
	defer mutex.Unlock()
	if res,ok := root.Find(sp); ok {
		if etagHandler, ok := res.(resource.ETagHandler); ok {
			if etagHandler.ETag() == eTag {
				root.UnMap(sp)
				return true
			}
		}
	}
	return false;
}


func Get(path string) (interface{}, bool) {
	sp := Standardize(path)
	mutex.Lock()
	defer mutex.Unlock()
	res, ok := root.Find(sp)
	return res,ok
}

func Has(path string) bool {
	_, ok := Get(path);
	return ok
}


func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sp := Standardize(r.URL.Path)
	var res interface{}

	if sp == "" {
		res = root
	} else {
		res, _ = root.Find(sp)
	}

	if (res != nil) {
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

