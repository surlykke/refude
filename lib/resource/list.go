// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package resource

import (
	"net/http"
	"strings"
	"sync"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

type ResourceCollection interface {
	CanServe(path string) bool
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Search(term string, threshold int) link.List
}

type Listener func() 

/**
* Behave like an ordered syncronized map
* Order is determined by insertion
 */
type Collection[T Resource] struct {
	sync.Mutex
	context   string
	resources []T
	listeners []Listener
}

func MakeCollection[T Resource](context string) *Collection[T] {
	return &Collection[T]{
		context:   context,
		resources: make([]T, 0, 100),
		listeners: []Listener{},
	}
}

func (this *Collection[T]) Get(id string) (T, bool) {
	this.Lock()
	defer this.Unlock()
	for _, res := range this.resources {
		if res.GetId() == id {
			return res, true
		}
	}
	var t T
	return t, false
}

func (this *Collection[T]) GetByPath(path string) (T, bool) {
	var id string
	if strings.HasPrefix(path, this.context) {
		id = path[len(this.context):]
	}
	return this.Get(id)
}


func (this *Collection[T]) GetAll() []T {
	this.Lock()
	defer this.Unlock()

	var list = make([]T, 0, len(this.resources))
	for _, res := range this.resources {
		list = append(list, res)
	}
	return list
}

func (this *Collection[T]) Search(term string, threshold int) link.List {
	if len(term) < threshold {
		return link.List{}
	}

	this.Lock()
	defer this.Unlock()
	var links = make(link.List, 0, len(this.resources))
	for _, res := range this.resources {
		if res.RelevantForSearch() {
			var title, _, icon, profile = res.Presentation()
			if rnk := searchutils.Match(term, title, res.GetKeywords()...); rnk > -1 {
				links = append(links, link.MakeRanked(res.GetId(), title, icon, profile, rnk))
			}
		}
	}
	links.SortByRank()
	return links
}

func (this *Collection[T]) Put(res T) {
	this.Lock()
	defer  this.publish()
	defer this.Unlock()
	for i := 0; i < len(this.resources); i++ {
		if this.resources[i].GetId() == res.GetId() {
			this.resources[i] = res
			return
		}
	}
	this.resources = append(this.resources, res)
}

func (this *Collection[T]) Update(res T) {
	this.Lock() 
	defer this.Unlock()
	for i := 0; i < len(this.resources); i++ {
		if this.resources[i].GetId() == res.GetId() {
			this.resources[i] = res
			this.publish()
			break
		}
	}
}

func (this *Collection[T]) PutFirst(res T) {
	this.Lock()
	defer this.publish()
	defer this.Unlock()
	for i := 0; i < len(this.resources); i++ {
		if this.resources[i].GetId() == res.GetId() {
			this.resources[i] = res
			return
		}
	}
	this.resources = append([]T{res}, this.resources...)
}

func (this *Collection[T]) ReplaceWith(resources []T) {
	this.Lock()
	this.resources = resources
	this.Unlock()
	this.publish()
}

func (this *Collection[T]) Delete(path string) bool {
	this.Lock()
	defer this.Unlock()

	for i, res := range this.resources {
		if res.GetId() == path {
			this.resources = append(this.resources[:i], this.resources[i+1:]...)
			this.publish()
			return true
		}
	}
	return false
}

func (this *Collection[T]) FindFirst(test func(t T) bool) (T, bool) {
	this.Lock()
	defer this.Unlock()

	for _, res := range this.resources {
		if test(res) {
			return res, true
		}
	}
	var t T
	return t, false
}

func (this *Collection[T]) Find(test func(t T) bool) []T {
	this.Lock()
	defer this.Unlock()

	var found = make([]T, 0, 5)

	for _, res := range this.resources {
		if test(res) {
			found = append(found, res)
		}
	}

	return found
}

func (this *Collection[T]) GetPaths() []string {
	var res = make([]string, 0, len(this.resources))
	for _, r := range this.resources {
		res = append(res, this.context + r.GetId())
	}
	return res
}

func (this *Collection[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == this.context {
		// Serve as list
		if r.Method == "GET" {
			var jsonReps = make([]jsonRepresentation, 0, 500)
			for _, res := range this.GetAll() {
				jsonReps = append(jsonReps, buildJsonRepresentation(res, this.context, ""))
			}
			respond.AsJson(w, jsonReps)
		} else {
			respond.NotAllowed(w)
		}
	} else if res, ok := this.GetByPath(r.URL.Path); ok { 
		ServeSingleResource(w, r, res, this.context)
	} else {
		respond.NotFound(w)
	}	
}

func (this *Collection[T]) AddListener(listener Listener) {
	this.listeners = append(this.listeners, listener)
}

func (this *Collection[T]) publish() {
	for _, listener := range this.listeners {
		listener()
	}
}
