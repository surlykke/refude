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
	less func(r1, r2 Resource) bool
}


func MakeOrderedCollection(less func(r1, r2 Resource) bool) *Collection {
	return &Collection {
		resources: make(map[string]Resource, 20),
		less: less,
	}
}

func MakeCollection() *Collection {
	return &Collection{
		resources: make(map[string]Resource, 20),
		less: defaultLess,
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
	sl := sortableList{resources: all, less: l.less}
	sort.Sort(&sl)
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

func defaultLess(r1, r2 Resource) bool {
	return r1.Self() < r2.Self()
}

/* ---------- Used by GetAll, so we have predictable order --------- */
type sortableList struct {
	less func(r1, r2 Resource) bool
	resources []Resource 
}

// Len is the number of elements in the collection.
func (sl *sortableList) Len() int {
	return len(sl.resources)
}

func (sl *sortableList) Less(i int, j int) bool {
	return sl.less(sl.resources[i], sl.resources[j])
}

func (sl *sortableList) Swap(i int, j int) {
	sl.resources[i], sl.resources[j] = sl.resources[j], sl.resources[i]	
}

