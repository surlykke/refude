// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"encoding/json"
	"fmt"
	"github.com/surlykke/RefudeServices/lib/parser"
	"github.com/surlykke/RefudeServices/lib/requests"
	"net/http"
	"strings"
	"sync"
)

type JsonResourceMap struct {
	mutex sync.Mutex
	rmap  map[StandardizedPath]*JsonResource
	links *JsonResource
}

func MakeJsonResourceMap() *JsonResourceMap {
	var m = &JsonResourceMap{mutex: sync.Mutex{}, rmap: make(map[StandardizedPath]*JsonResource), links: nil}
	return m
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

func (jm *JsonResourceMap) unput(sp StandardizedPath) (Resource, bool) {
	if reservedPaths[sp] {
		panic("Attempt to unmap reserved path: " + sp)
	}

	if jsonRes, ok := jm.rmap[sp]; ok {
		delete(jm.rmap, sp)
		jm.links = nil
		return jsonRes.GetRes(), true
	} else {
		return nil, false
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

func (jm *JsonResourceMap) Unmap(path StandardizedPath) (Resource, bool) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	return jm.unput(path)
}

// --------------------------- New api ----------------------------------------------------------

type Mapping struct {
	Path     StandardizedPath
	Resource Resource
}

type Update struct {
	PathsToRemove    []StandardizedPath
	PrefixesToRemove []StandardizedPath
	Mappings         []Mapping
}

func (jm *JsonResourceMap) Run(updateStream <- chan Update) {
	for update := range updateStream {
		jm.doUpdate(update)
	}
}

func (jm *JsonResourceMap) doUpdate(update Update) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	for _, pathToRemove := range update.PathsToRemove {
		jm.unput(pathToRemove)
	}

	for _, prefixToRemove := range update.PrefixesToRemove {
		for path, _ := range jm.rmap {
			if strings.HasPrefix(string(path), string(prefixToRemove)) {
				jm.unput(path)
			}
		}
	}

	for _,mapping := range update.Mappings {
		jm.put(mapping.Path, mapping.Resource)
	}
}

// --------------------------- Implement JsonCollection -----------------------------------------

func (jm *JsonResourceMap) GetResource(path StandardizedPath) *JsonResource {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	var res, ok = jm.rmap[path]
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
		result[pos] = res
		res.EnsureReady()
		pos++
	}
	return result
}

func (jm *JsonResourceMap) ensureLinksUpdated() {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	if jm.links == nil {
		var mtMap = make(Links)
		mtMap["application/json"] = []StandardizedPath{"/links", "/search"}
		for sp, res := range jm.rmap {
			mtMap[res.GetMt()] = append(mtMap[res.GetMt()], sp)
		}

		jm.links = MakeJsonResource(&mtMap)
		jm.links.EnsureReady()
	}
}

// ----------------------------------------------------------------------------------------------

func (jm *JsonResourceMap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sp := Standardize(r.URL.Path)
	if sp == "/links" {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
		} else {
			jm.ensureLinksUpdated()
			jm.links.ServeHTTP(w, r)
		}
	} else if sp == "/search" {
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
	} else {
		if res := jm.GetResource(sp); res != nil {
			res.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func (jm *JsonResourceMap) findResources(mt MediaType, query string) ([]*JsonResource, error) {
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
}
