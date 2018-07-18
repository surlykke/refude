// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package service

import (
	"github.com/surlykke/RefudeServices/lib/resource"
	"strings"
	"fmt"
	"sync"
)
// A standardized path is a path that starts with '/' and has no double slashes
type StandardizedPath string

type JsonResourceMap struct {
	mutex sync.Mutex
	rmap  map[StandardizedPath]*resource.JsonResource
}


func MakeJsonResourceMap() *JsonResourceMap {
	return &JsonResourceMap{rmap: make(map[StandardizedPath]*resource.JsonResource)}
}


/** transform a path to a standardized path
 * Watered down version of path.Clean. Replace any sequence of '/' with single '/'
 * Remove ending '/'
 * We do not resolve '..', (so '/foo/../baa' is different from '/baa')
 * Examples:
 *       '//foo/baa' becomes '/foo/baa'
 *       '/foo///baa/////muh/' becomes '/foo/baa/muh'
 *       '/foo/..//baa//' becomes '/foo/../baa'
 */
func standardize(p string) StandardizedPath {
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
		return StandardizedPath(buffer[:pos-1])
	} else {
		return StandardizedPath(buffer[:pos])
	}

}

/**
	Break standardized path into dir-part and base-part
    '/foo/baa/res' -> '/foo/baa', 'res'
    '/foo/baa' -> '/foo', 'baa'
 */
func separate(sp StandardizedPath) (StandardizedPath, string) {
	if len(sp) == 0 {
		panic("Separating empty string")
	} else {
		var pos = strings.LastIndexByte(string(sp[:len(sp)-1]), '/')
		return sp[:pos], string(sp[pos+1:])
	}
}



var reservedPaths = map[StandardizedPath]bool{
	"/links":  true,
	"/search": true,
}


// The public methods, further down, are responsible for thread safety - ie. aquiring of locks

func (jm *JsonResourceMap) put(sp StandardizedPath, res resource.Resource) {
	if reservedPaths[sp] {
		panic("Attempt to map to reserved path: " + sp)
	}
	var jsonResource = resource.MakeJsonResource(res)
	jm.rmap[sp] = jsonResource
	delete(jm.rmap, "/links")
}

func (jm *JsonResourceMap) unput(sp StandardizedPath) (resource.Resource, bool) {
	if reservedPaths[sp] {
		panic("Attempt to unmap reserved path: " + sp)
	}

	if jsonRes, ok := jm.rmap[sp]; ok {
		delete(jm.rmap, sp)
		delete(jm.rmap, "/links")
		return jsonRes.GetRes(), true
	} else {
		return nil, false
	}
}

// ------------------------------------ Public ----------------------------------------------------

func (jm *JsonResourceMap) RemoveAll(dirpath string) {
	var lookFor = string(standardize(dirpath) + "/")
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	for path, _ := range jm.rmap {
		if strings.HasPrefix(string(path), lookFor) {
			jm.unput(path)
		}
	}
}

func (jm *JsonResourceMap) Map(res resource.Resource) {
	if self := res.GetSelf(); self == "" {
		panic("Mapping resource with empty self")
	} else {
		sp := standardize(self)
		jm.mutex.Lock()
		defer jm.mutex.Unlock()
		jm.put(sp, res)
	}
}

func (jm *JsonResourceMap) Unmap(path string) (resource.Resource, bool ){
	sp := standardize(path)
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	return jm.unput(sp)
}

// --------------------------- Implement JsonCollection -----------------------------------------

func (jm *JsonResourceMap) GetResource(path StandardizedPath) *resource.JsonResource {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	if (path == "/links") {
		if _,ok := jm.rmap["/links"]; !ok {
			var links = make(Links);
			for path, jsonRes := range jm.rmap {
				links[jsonRes.GetMt()] = append(links[jsonRes.GetMt()], path)
			}
			jm.rmap["/links"] = resource.MakeJsonResource(links)
		}
	}
	var res, ok = jm.rmap[path];
	if ok {
		res.EnsureReady()
	}
	return res
}




func (jm *JsonResourceMap) GetAll() []*resource.JsonResource {
	var result = make([]*resource.JsonResource, len(jm.rmap))
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	var pos = 0
	for _, res := range jm.rmap {
		result[pos]	= res
		res.EnsureReady()
		pos++
	}
	return result
}

// ----------------------------------------------------------------------------------------------


//func ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	sp := standardize(r.URL.Path)
//	if sp == "/search" {
//		if r.Method == "GET" {
//			Search(w, r)
//		} else {
//			w.WriteHeader(http.StatusMethodNotAllowed)
//		}
//	} else if sp == "/links" {
//		links.ServeHTTP(w,r)
//	} else if res, ok := findForServing(sp); !ok {
//		w.WriteHeader(http.StatusNotFound)
//	} else {
//		res.ServeHTTP(w, r)
//	}
//}


