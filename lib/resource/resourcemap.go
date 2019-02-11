// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"net/http"
	"strings"
	"sync"
)

type JsonResourceMap struct {
	mutex sync.Mutex
	rmap  map[StandardizedPath]http.Handler
	links *JsonResource
}

func MakeJsonResourceMap() *JsonResourceMap {
	var m = &JsonResourceMap{mutex: sync.Mutex{}, rmap: make(map[StandardizedPath]http.Handler), links: nil}
	return m
}

type Mapping struct {
	Path     StandardizedPath
	Resource Resource
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
	jm.links = nil
}

func (jm *JsonResourceMap) unput(sp StandardizedPath) {
	if reservedPaths[sp] {
		panic("Attempt to unmap reserved path: " + sp)
	}

	if _, ok := jm.rmap[sp]; ok {
		delete(jm.rmap, sp)
		jm.links = nil
	}
}

// ------------------------------------ Public ----------------------------------------------------

func (jm *JsonResourceMap) RemoveAll(prefixes ...StandardizedPath) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	for _, prefix := range prefixes {
		for path, _ := range jm.rmap {
			if strings.HasPrefix(string(path), string(prefix)) {
				jm.unput(path)
			}
		}
	}
}

func (jm *JsonResourceMap) RemoveAndMap(prefixesToRemove []StandardizedPath, resources []Resource) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	for path, _ := range jm.rmap {
		for _, prefix := range prefixesToRemove {
			if strings.HasPrefix(string(path), string(prefix)) {
				jm.unput(path)
				break
			}
		}
	}

	for _, resource := range resources {
		jm.put(resource.GetSelf(), resource)
	}
}

func (jm *JsonResourceMap) Map(res Resource) {
	if res.GetSelf() == "" {
		panic("Mapping resource with empty self")
	} else {
		jm.MapTo(res.GetSelf(), res)
	}
}

func (jm *JsonResourceMap) MapTo(path StandardizedPath, res Resource) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	jm.put(path, res)
}

func (jm *JsonResourceMap) Update(prefixesToRemove []StandardizedPath, mappings []Mapping) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	for _, prefixToRemove := range prefixesToRemove {
		for path, _ := range jm.rmap {
			if strings.HasPrefix(string(path), string(prefixToRemove)) {
				jm.unput(path)
			}
		}
	}

	for _, mapping := range mappings {
		jm.put(mapping.Path, mapping.Resource)
	}
}

func (jm *JsonResourceMap) Unmap(path StandardizedPath) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	jm.unput(path)
}

// --------------------------- Implement JsonCollection -----------------------------------------

func (jm *JsonResourceMap) GetResource(path StandardizedPath) http.Handler {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	return jm.rmap[path]
}

func (jm *JsonResourceMap) GetAll() []http.Handler {
	var result = make([]http.Handler, len(jm.rmap))
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	var pos = 0
	for _, res := range jm.rmap {
		result[pos] = res
		pos++
	}
	return result
}

// ----------------------------------------------------------------------------------------------

func (jm *JsonResourceMap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sp := Standardize(r.URL.Path)
	if sp == "/links" {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
		} else {
			jm.links.ServeHTTP(w, r)
		}
	} else /*if sp == "/search" {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
		} else if flatParams, err := requests.GetSingleParams(r, "type", "q"); err != nil {
			requests.ReportUnprocessableEntity(w, err)
		} else if resources, err := jm.findResources(MediaType(flatParams["type"]), flatParams["q"]); err != nil {
			requests.ReportUnprocessableEntity(w, err)
		} else if bytes, err := json.Marshal(resources); err != nil {
			panic(fmt.Sprintln("Unable to marshall search result", err))
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(bytes)
		}
	} else */{
		if res := jm.GetResource(sp); res != nil {
			res.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

/*func (jm *JsonResourceMap) findResources(mt MediaType, query string) ([]*JsonResource, error) {
	var tmp = jm.GetAll()
	var found = 0;
	if mt != "" {
		for i := 0; i < len(tmp); i++ {
			if MediaTypeMatch(MediaType(mt), tmp[i].GetMt()) {
				tmp[found] = tmp[i]
				found++
			}
		}
		tmp = tmp[:found]
	}

	if query != "" {
		if matcher, err := parser.Parse(query); err != nil {
			return nil, err
		} else {
			found = 0;
			for i := 0; i < len(tmp); i++ {
				if matcher(tmp[i].GetRes()) {
					tmp[found] = tmp[i]
					found++
				}
			}
			tmp = tmp[:found]
		}
	}

	return tmp, nil
}*/
