// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package repo

import (
	"cmp"
	"fmt"
	"sync"

	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/internal/lib/link"
	"github.com/surlykke/refude/internal/lib/response"
)

type SyncMap[K cmp.Ordered, V entity.Servable] struct {
	m        map[K]V
	lock     sync.Mutex
	basepath string
}

func MakeSynkMap[K cmp.Ordered, V entity.Servable]() *SyncMap[K, V] {
	var m = &SyncMap[K, V]{
		m: make(map[K]V),
	}

	return m
}

// Called before trafic begins
func (this *SyncMap[K, V]) SetPrefix(basepath string) {
	this.basepath = basepath
	this.setPaths()
}

func (this *SyncMap[K, V]) Get(k K) (V, bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
	v, ok := this.m[k]
	return v, ok
}

func (this *SyncMap[K, V]) Put(k K, v V) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.put(k, v)
}

func (this *SyncMap[K, V]) put(k K, v V) {
	v.GetBase().Path = fmt.Sprintf("%s%v", this.basepath, k)
	v.GetBase().BuildLinks()
	this.m[k] = v
}

func (this *SyncMap[K, V]) Remove(k K) (V, bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.remove(k)
}

func (this *SyncMap[K, V]) remove(k K) (V, bool) {
	v, ok := this.m[k]
	if ok {
		delete(this.m, k)
	}
	return v, ok
}

func (this *SyncMap[K, V]) Replace(newVals map[K]V, remove func(V) bool) {
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

func (this *SyncMap[K, V]) ReplaceAll(newSet map[K]V) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.m = newSet
	this.setPaths()
}

func (this *SyncMap[K, V]) GetAll() []V {
	this.lock.Lock()
	defer this.lock.Unlock()
	var list = make([]V, 0, len(this.m))
	for _, v := range this.m {
		list = append(list, v)
	}
	return list
}

func (this *SyncMap[K, V]) GetForSearch() []entity.Base {
	var bases = make([]entity.Base, 0, len(this.m))
	for _, v := range this.GetAll() {
		if v.OmitFromSearch() {
			continue
		}
		bases = append(bases, *v.GetBase())
	}
	return bases
}

func (this *SyncMap[K, V]) DoGetSingle(id K) response.Response {
	if v, ok := this.Get(id); ok {
		return response.Json(v)
	} else {
		return response.NotFound()
	}
}

func (this *SyncMap[K, V]) DoGetAll() response.Response {
	return response.Json(this.GetAll())
}

func (this *SyncMap[K, V]) DoPost(id K, action string) response.Response {
	if v, ok := this.Get(id); !ok {
		return response.NotFound()
	} else if postable, ok := any(v).(link.Postable); !ok {
		return response.NotAllowed()
	} else {
		return postable.DoPost(action)
	}
}

func (this *SyncMap[K, V]) GetPaths() []string {
	var paths = make([]string, 0, len(this.m))
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, v := range this.m {
		paths = append(paths, v.GetBase().Path)
	}
	return paths
}

func (this *SyncMap[K, V]) setPaths() {
	for k, v := range this.m {
		v.GetBase().Path = fmt.Sprintf("%s%v", this.basepath, k)
		v.GetBase().BuildLinks()
	}
}
