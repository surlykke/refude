// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package entity

import (
	"cmp"
	"fmt"
	"sync"

	"github.com/surlykke/refude/internal/lib/response"
)

type EntityMap[K cmp.Ordered, V Servable] struct {
	m        map[K]V
	lock     sync.Mutex
	basepath string
}

func MakeMap[K cmp.Ordered, V Servable]() *EntityMap[K, V] {
	var m = &EntityMap[K, V]{
		m: make(map[K]V),
	}

	return m
}

// Called before trafic begins
func (this *EntityMap[K, V]) SetPrefix(basepath string) {
	this.basepath = basepath
	this.setPaths()
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
}

func (this *EntityMap[K, V]) put(k K, v V) {
	v.GetBase().Path = fmt.Sprintf("%s%v", this.basepath, k)
	v.GetBase().BuildLinks()
	this.m[k] = v
}

func (this *EntityMap[K, V]) Remove(k K) (V, bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.remove(k)
}

func (this *EntityMap[K, V]) remove(k K) (V, bool) {
	v, ok := this.m[k]
	if ok {
		delete(this.m, k)
	}
	return v, ok
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
}

func (this *EntityMap[K, V]) ReplaceAll(newSet map[K]V) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.m = newSet
	this.setPaths()
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

func (this *EntityMap[K, V]) DoGetSingle(id K) response.Response {
	if v, ok := this.Get(id); ok {
		return response.Json(v)
	} else {
		return response.NotFound()
	}
}

func (this *EntityMap[K, V]) DoGetAll() response.Response {
	return response.Json(this.GetAll())
}

func (this *EntityMap[K, V]) DoPost(id K, action string) response.Response {
	if v, ok := this.Get(id); !ok {
		return response.NotFound()
	} else if postable, ok := any(v).(Postable); !ok {
		return response.NotAllowed()
	} else {
		return postable.DoPost(action)
	}
}

func (this *EntityMap[K, V]) GetPaths() []string {
	var paths = make([]string, 0, len(this.m))
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, v := range this.m {
		paths = append(paths, v.GetBase().Path)
	}
	return paths
}

func (this *EntityMap[K, V]) setPaths() {
	for k, v := range this.m {
		v.GetBase().Path = fmt.Sprintf("%s%v", this.basepath, k)
		v.GetBase().BuildLinks()
	}
}
