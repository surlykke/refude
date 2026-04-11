// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package entity

import (
	"cmp"
	"fmt"
	"net/http"
	"sync"

	"github.com/surlykke/refude/internal/lib/respond"
	"github.com/surlykke/refude/internal/lib/utils"
	"github.com/surlykke/refude/pkg/pubsub"
)

type Event struct {
	Event string
	Data  string
}

type EntityMap[K cmp.Ordered, V Servable] struct {
	m      map[K]V
	lock   sync.Mutex
	Prefix string
	Events *pubsub.Publisher[Event]
}

func MakeMap[K cmp.Ordered, V Servable](prefix string) *EntityMap[K, V] {
	var m = &EntityMap[K, V]{
		m:      make(map[K]V),
		Prefix: prefix,
		Events: pubsub.MakePublisher[Event](),
	}

	return m
}

func (this *EntityMap[K, V]) Get(k K) (V, bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
	v, ok := this.m[k]
	return v, ok
}

func (this *EntityMap[K, V]) Put(k K, v V) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.put(k, v)
	this.publishWithId(k)
}

func (this *EntityMap[K, V]) Remove(k K) (V, bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.publishWithId(k)
	return this.remove(k)
}

func (this *EntityMap[K, V]) Replace(newVals map[K]V, remove func(V) bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
	for k, v := range this.m {
		if remove(v) {
			delete(this.m, k)
		}
	}
	for k, v := range newVals {
		this.put(k, v)
	}
	this.publish()
}

func (this *EntityMap[K, V]) ReplaceAll(newSet map[K]V) {
	for id, v := range newSet {
		v.GetBase().SetPath(fmt.Sprintf("%s%v", this.Prefix, id))
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	this.m = newSet
	this.publish()
}

func (this *EntityMap[K, V]) GetAll() []V {
	this.lock.Lock()
	defer this.lock.Unlock()
	var list = make([]V, 0, len(this.m))
	for _, v := range this.m {
		list = append(list, v)
	}
	return list
}

func (this *EntityMap[K, V]) GetForSearch() []Base {
	var bases = make([]Base, 0, len(this.m))
	for _, v := range this.GetAll() {
		if v.OmitFromSearch() {
			continue
		}
		bases = append(bases, *v.GetBase())
	}
	return bases
}

func (this *EntityMap[K, V]) GetPaths() []string {
	var paths = make([]string, 0, len(this.m))
	this.lock.Lock()
	defer this.lock.Unlock()
	for id, _ := range this.m {
		paths = append(paths, fmt.Sprintf("%s%v", this.Prefix, id))
	}
	return paths
}

func (this *EntityMap[K, V]) put(k K, v V) {
	v.GetBase().SetPath(fmt.Sprintf("%s%v", this.Prefix, k))
	this.m[k] = v
}

func (this *EntityMap[K, V]) remove(k K) (V, bool) {
	v, ok := this.m[k]
	if ok {
		delete(this.m, k)
	}
	return v, ok
}

func (this *EntityMap[K, V]) publishWithId(id K) {
	this.Events.Publish(Event{Event: this.Prefix, Data: fmt.Sprintf("%v", id)})
}

func (this *EntityMap[K, V]) publish() {
	this.Events.Publish(Event{Event: this.Prefix, Data: ""})
}

func (this *EntityMap[K, V]) Serve() {
	http.HandleFunc("GET "+this.Prefix+"{id...}", func(w http.ResponseWriter, r *http.Request) {
		var id = r.PathValue("id")
		if id == "" {
			respond.AsJson(w, this.GetAll())
		} else if v, ok := this.GetByStr(id); !ok {
			respond.NotFound(w)
		} else {
			respond.AsJson(w, v)
		}
	})
	http.HandleFunc("POST "+this.Prefix+"{id...}", func(w http.ResponseWriter, r *http.Request) {
		if v, ok := this.GetByStr(r.PathValue("id")); !ok {
			respond.NotFound(w)
		} else if postable, ok := any(v).(Postable); !ok {
			respond.NotAllowed(w)
		} else {
			if ok, err := postable.DoPost(utils.QueryParam(r, "action")); err != nil {
				respond.ServerError(w, err)
			} else if !ok {
				respond.NotFound(w)
			} else {
				respond.Accepted(w)
			}
		}
	})
	http.HandleFunc("DELETE "+this.Prefix+"{id...}", func(w http.ResponseWriter, r *http.Request) {
		if v, ok := this.GetByStr(r.PathValue("id")); !ok {
			respond.NotFound(w)
		} else if deleteable, ok := any(v).(Deleteable); !ok {
			respond.NotAllowed(w)
		} else if err := deleteable.DoDelete(); err != nil {
			respond.ServerError(w, err)
		} else {
			respond.Accepted(w)
		}
	})

}

func (this *EntityMap[K, V]) GetByStr(idStr string) (V, bool) {
	var id K
	if err := utils.Convert(idStr, &id); err != nil {
		var zeroval V
		return zeroval, false
	} else {
		return this.Get(id)
	}
}
