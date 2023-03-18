// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package resource

import (
	"sync"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

/**
 * Behave like an ordered syncronized map
 * Order is determined by insertion
 */
type Collection[T Resource] struct {
	sync.Mutex
	resources []T
}

func MakeCollection[T Resource]() *Collection[T] {
	return &Collection[T]{
		resources: make([]T, 0, 100),
	}
}


func (l *Collection[T]) Get(path string) (T, bool) {
	l.Lock()
	defer l.Unlock()
	for _, res := range l.resources {
		if res.GetPath() == path  {
			return res, true
		}
	}
	var t T
	return t, false
}

func (l *Collection[T]) GetAll() []T {
	l.Lock()
	defer l.Unlock()

	var list = make([]T, 0, len(l.resources))
	for _, res := range l.resources {
		list = append(list, res)
	}
	return list
}


func (l *Collection[T]) GetResource(path string) Resource {
	if res, ok := l.Get(path); ok {
		return res
	} else {
		return nil
	}
}

func (l *Collection[T]) GetResources() []Resource {
	l.Lock()
	defer l.Unlock()
	var all = make([]Resource, 0, len(l.resources))
	for _, res := range l.resources {
		all = append(all, res)
	}
	return all
}

func (l *Collection[T]) Search(term string, threshold int) link.List {
	if len(term) < threshold {
		return link.List{}
	}

	l.Lock()
	defer l.Unlock()
	var links = make(link.List, 0, len(l.resources))
	for _, res := range l.resources {
		if res.RelevantForSearch() {
			var title, _, icon, profile = res.Presentation() 
			if rnk := searchutils.Match(term, title, res.GetKeywords()...); rnk > -1 {
				links = append(links, link.MakeRanked(res.GetPath(), title, icon, profile, rnk))
			}
		}
	}
	links.SortByRank()
	return links
}


func (l *Collection[T]) Put(res T) {
	l.Lock()
	defer l.Unlock()
	for i := 0; i < len(l.resources); i++ {
		if l.resources[i].GetPath() == res.GetPath() {
			l.resources[i] = res
			return
		}
	}
	l.resources = append(l.resources, res)
}

func (l *Collection[T]) ReplaceWith(resources []T) {
	l.Lock()
	defer l.Unlock()
	l.resources = resources
}

func (l *Collection[T]) Delete(path string) bool {
	l.Lock()
	defer l.Unlock()

	for i, res := range l.resources {
		if res.GetPath() == path {
			l.resources = append(l.resources[:i], l.resources[i+1:]...)
			return true
		}
	}
	return false
}

func (l *Collection[T]) FindFirst(test func(t T) bool) (T, bool) {
	l.Lock()
	defer l.Unlock()

	for _, res := range l.resources {
		if test(res) {
			return res, true
		}
	}
	var t T
	return t, false
}



func (l *Collection[T]) GetPaths() []string {
	var res = make([]string, 0, len(l.resources))
	for _, r := range l.resources {
		res = append(res, r.GetPath())
	}
	return res
}


