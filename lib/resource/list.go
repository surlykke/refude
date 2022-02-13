// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"sort"
	"sync"
)

/**
	Basically, a syncronized map
*/
type Collection struct {
	sync.Mutex
	Prefix    string
	resources map[string]Resource
}


func MakeCollection() *Collection {
	return &Collection{
		resources: make(map[string]Resource, 20),
	}
}

func (l *Collection) Get(path string) Resource {
	l.Lock()
	defer l.Unlock()
	return l.resources[path]
}


func (l *Collection) GetAll() []Resource {
	l.Lock()
	defer l.Unlock()
	var all = make([]Resource,0, len(l.resources))
	for _, res := range l.resources {
		all = append(all, res)
	}
	sort.Sort(sortableList(all))
	return all
}

func (l *Collection) Put(res Resource) {
	l.Lock()
	defer l.Unlock()
	l.handlePrefix(res.Self())
	l.resources[res.Self()] = res
}


func (l *Collection) handlePrefix(path string) {
	if len(path) == 0 || path[0] != '/' {
		panic("Paths should start with '/'" + path)
	} else {
		for i := 1; i < len(path); i++ {
			if path[i] == '/' {
				var prefix = path[:i+1]
				if len(prefix) == 2 {
					panic("Thats not a serious path prefix: " + prefix)
				}
				if l.Prefix == "" {
					l.Prefix = prefix
				} else if l.Prefix != prefix {
					panic("All resources added to collection should have same prefix: " + l.Prefix + "," + prefix)
				}
				return
			}
		}
		panic("Resource path does not have a prefix: " + path)
	}
}

func (l *Collection) ReplaceWith(resources []Resource) {
	l.Lock()
	defer l.Unlock()
	
	l.resources = make(map[string]Resource, len(resources))
	for _, res := range resources {
		l.handlePrefix(res.Self())
		l.resources[res.Self()] = res
	}
}

func (l *Collection) Delete(path string) bool {
	l.Lock()
	defer l.Unlock()

	if _, ok := l.resources[path]; ok {
		delete(l.resources, path)	
		return true
	} else {
		return false
	}
}

func (l *Collection) FindFirst(test func(res Resource) bool) interface{} {
	l.Lock()
	defer l.Unlock()
	for _, res := range l.resources {
		if test(res) {
			return res
		}
	}
	return nil
}

/* ---------- Used by GetAll, so we have predictable order --------- */
type sortableList []Resource 

// Len is the number of elements in the collection.
func (rl sortableList) Len() int {
	return len(rl)
}

func (rl sortableList) Less(i int, j int) bool {
	return rl[i].Self() < rl[j].Self()
}

func (rl sortableList) Swap(i int, j int) {
	rl[i], rl[j] = rl[j], rl[i]	
}

