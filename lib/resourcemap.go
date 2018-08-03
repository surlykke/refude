// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package lib

import (
	"strings"
	"sync"
)

type JsonResourceMap struct {
	mutex sync.Mutex
	rmap  map[StandardizedPath]*JsonResource
}


func MakeJsonResourceMap() *JsonResourceMap {
	return &JsonResourceMap{rmap: make(map[StandardizedPath]*JsonResource)}
}




var reservedPaths = map[StandardizedPath]bool{
	"/links":  true,
	"/search": true,
}


// The public methods, further down, are responsible for thread safety - ie. aquiring of locks

func (jm *JsonResourceMap) put(sp StandardizedPath, res Resource) {
	if reservedPaths[sp] {
		panic("Attempt to map to reserved path: " + sp)
	}
	var jsonResource = MakeJsonResource(res)
	jm.rmap[sp] = jsonResource
	delete(jm.rmap, "/links")
}

func (jm *JsonResourceMap) unput(sp StandardizedPath) (Resource, bool) {
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
	var lookFor = string(Standardize(dirpath) + "/")
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	for path, _ := range jm.rmap {
		if strings.HasPrefix(string(path), lookFor) {
			jm.unput(path)
		}
	}
}

func (jm *JsonResourceMap) RemoveAndMap(prefixesToRemove []string, resources []Resource) {
	var prefixesStandardized = make([]StandardizedPath, len(prefixesToRemove))
	for i,pr := range prefixesToRemove {
		prefixesStandardized[i] = Standardize(pr)
	}
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	for path,_ := range jm.rmap {
		for _,prefix := range prefixesStandardized {
			if strings.HasPrefix(string(path), string(prefix)) {
				jm.unput(path)
				break
			}
		}
	}

	for _,resource := range resources {
		jm.put(resource.GetSelf(), resource)
	}
}

func (jm *JsonResourceMap) Map(res Resource) {
	if res.GetSelf() == "" {
		panic("Mapping resource with empty self")
	} else {
		jm.mutex.Lock()
		defer jm.mutex.Unlock()
		jm.put(res.GetSelf(), res)
	}
}

func (jm *JsonResourceMap) Unmap(path StandardizedPath) (Resource, bool ){
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	return jm.unput(path)
}

// --------------------------- Implement JsonCollection -----------------------------------------

func (jm *JsonResourceMap) GetResource(path StandardizedPath) *JsonResource {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	if (path == "/links") {
		if _,ok := jm.rmap["/links"]; !ok {
			var links = make(Links);
			for path, jsonRes := range jm.rmap {
				links[jsonRes.GetMt()] = append(links[jsonRes.GetMt()], path)
			}
			jm.rmap["/links"] = MakeJsonResource(links)
		}
	}
	var res, ok = jm.rmap[path];
	if ok {
		res.EnsureReady()
	}
	return res
}




func (jm *JsonResourceMap) GetAll() []*JsonResource {
	var result = make([]*JsonResource, len(jm.rmap))
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


