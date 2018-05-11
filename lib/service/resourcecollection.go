// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package service

import (
	"net/http"
	"github.com/surlykke/RefudeServices/lib/resource"
	"strings"
	"fmt"
	"sync"
	"github.com/surlykke/RefudeServices/lib/mediatype"
)

// A standardized path is a path that starts with '/' and has no double slashes
type standardizedPath string

/** transform a path to a standardized path
 * Watered down version of path.Clean. Replace any sequence of '/' with single '/'
 * Remove ending '/'
 * We do not resolve '..', (so '/foo/../baa' is different from '/baa')
 * Examples:
 *       '//foo/baa' becomes '/foo/baa'
 *       '/foo///baa/////muh/' becomes '/foo/baa/muh'
 *       '/foo/..//baa//' becomes '/foo/../baa'
 */
func standardize(p string) standardizedPath {
	if len(p) == 0 || p[0] != '/' {
		panic(fmt.Sprintf("path must start with '/': '%s'", p))
	}

	var buffer = make([]byte, len(p), len(p))
	var pos = 0
	var justSawSlash = false

	for i := 0; i < len(p); i++ {
		if !justSawSlash || p[i] != '/' {
			buffer[pos] = p[i]
			pos++
		}
		justSawSlash = p[i] == '/'
	}

	if buffer[pos-1] == '/' {
		return standardizedPath(buffer[:pos-1])
	} else {
		return standardizedPath(buffer[:pos])
	}

}

/**
	Break standardized path into dir-part and base-part
    '/foo/baa/res' -> '/foo/baa', 'res'
    '/foo/baa' -> '/foo', 'baa'
 */
func separate(sp standardizedPath) (standardizedPath, string) {
	if len(sp) == 0 {
		panic("Separating empty string")
	} else {
		var pos = strings.LastIndexByte(string(sp[:len(sp)-1]), '/')
		return sp[:pos], string(sp[pos+1:])
	}
}

var mutex sync.Mutex
var rc = make(map[standardizedPath]resource.Resource)

var links = MakeLinks()

var reservedPaths = map[standardizedPath]bool{
	"/links":  true,
	"/search": true,
}

func init() {
	rc["/links"] = links
}

// TODO: Note about threadsafety

func put(sp standardizedPath, res resource.Resource) {
	if reservedPaths[sp] {
		panic("Attempt to map to reserved path: " + sp)
	}
	rc[sp] = res
	links.addLinkEntry(sp, res.Mt())
	clearSearchCache()
}

func unput(sp standardizedPath) (resource.Resource, bool){
	if reservedPaths[sp] {
		panic("Attempt to unmap reserved path: " + sp)
	}

	if res, ok := rc[sp]; ok {
		delete(rc, sp)
		links.removeLinkEntry(sp)
		clearSearchCache()
		return res, true
	} else {
		return nil, false
	}
}

func findForServing(path standardizedPath) (resource.Resource, bool) {
	mutex.Lock()
	defer mutex.Unlock()
	var res, ok = rc[path];
	return res, ok
}

// ------------------------------------ Public ----------------------------------------------------

func RemoveAll(dirpath string) {
	var lookFor = string(standardize(dirpath) + "/")
	mutex.Lock()
	defer mutex.Unlock()
	for path, _ := range rc {
		if strings.HasPrefix(string(path), lookFor) {
			unput(path)
		}
	}
}

func MapAll(newEntries map[string]resource.Resource) {
	mutex.Lock()
	defer mutex.Unlock()
	for path, res := range newEntries {
		sp := standardize(path)
		put(sp, res)
	}

}

func Map(path string, mappableType resource.MappableType, mt mediatype.MediaType) {
	sp := standardize(path)
	mappableType.SetSelf(string(sp))
	var jsonResource = resource.MakeJsonResource(mappableType, mt)
	mutex.Lock()
	defer mutex.Unlock()
	put(sp, jsonResource)
}

func Unmap(path string) (resource.Resource, bool ){
	sp := standardize(path)
	mutex.Lock()
	defer mutex.Unlock()
	return unput(sp)
}

func Has(path string) bool {
	sp := standardize(path)
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := rc[sp]; ok {
		return true
	} else {
		return false
	}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sp := standardize(r.URL.Path)
	if sp == "/search" {
		if r.Method == "GET" {
			Search(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	} else if res, ok := findForServing(sp); !ok {
		w.WriteHeader(http.StatusNotFound)
	} else {
		res.ServeHTTP(w, r)
	}
}


