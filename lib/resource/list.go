// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

/**
 * A syncronized map, ordered by key
 */
type Collection[ID constraints.Ordered, T Resource[ID]] struct {
	sync.Mutex
	Prefix    string // Immutable
	resources map[ID]T
}

func MakeCollection[ID constraints.Ordered, T Resource[ID]](prefix string) *Collection[ID, T] {
	return &Collection[ID, T]{
		Prefix:    prefix,
		resources: make(map[ID]T),
	}
}

func (l *Collection[ID, T]) Get(id ID) T {
	l.Lock()
	defer l.Unlock()
	return l.resources[id]
}

func (l *Collection[ID, T]) Put(res T) {
	l.Lock()
	defer l.Unlock()

	l.resources[res.Id()] = res
}

func (l *Collection[ID, T]) GetAll() []T {
	l.Lock()
	defer l.Unlock()

	var all = make([]T, 0, len(l.resources))
	for _, res := range l.resources {
		all = append(all, res)
	}

	slices.SortFunc(all, func(t1, t2 T) bool { return t1.Id() < t2.Id() })

	return all
}

func (l *Collection[ID, T]) ReplaceWith(resources []T) {
	l.Lock()
	defer l.Unlock()

	l.resources = make(map[ID]T, len(resources))
	for _, res := range resources {
		l.resources[res.Id()] = res
	}
}

func (l *Collection[ID, T]) Delete(id ID) bool {
	l.Lock()
	defer l.Unlock()

	if _, ok := l.resources[id]; ok {
		delete(l.resources, id)
		return true
	} else {
		return false
	}
}

func (l *Collection[ID, T]) FindFirst(test func(t T) bool) (T, bool) {
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

func (l *Collection[ID, T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == l.Prefix {
		l.Lock()
		var links = make([]link.Link, 0, len(l.resources))
		for _, res := range l.resources {
			links = append(links, LinkTo[ID](res, l.Prefix, 0))
		}
		l.Unlock()
		respond.AsJson(w, links)
	} else if strings.HasPrefix(r.URL.Path, l.Prefix) {
		l.Lock()

		for id, res := range l.resources {
			var self = l.Self(id)
			if r.URL.Path == self {
				l.Unlock()
				ServeResource[ID](w, r, self, res)
				return
			}
		}
		l.Unlock()
		respond.NotFound(w)
	}
}

func (l *Collection[ID, T]) GetPaths() []string {
	var res = make([]string, 0, len(l.resources))
	for _, r := range l.resources {
		res = append(res, fmt.Sprint(l.Prefix, r.Id()))
	}
	return res
}

func (c *Collection[ID, T]) ExtractLinks(rank func(t T) int) link.List {
	var links = make(link.List, 0, len(c.resources))
	for _, t := range c.GetAll() {
		if rnk := rank(t); rnk > -1 {
			links = append(links, LinkTo[ID](t, c.Prefix, rnk))
		}
	}
	return links
}

func (l *Collection[ID, T]) Self(id ID) string {
	return fmt.Sprint(l.Prefix, id)
}


func ServeResource[ID constraints.Ordered, T Resource[ID]](w http.ResponseWriter, r *http.Request, self string, res T) {
	var linkSearchTerm = requests.GetSingleQueryParameter(r, "search", "")
	if r.Method == "GET" {
		respond.AsJson(w, MakeWrapper[ID](self, res, linkSearchTerm))
	} else {
		var resI Resource[ID] = res
		if postable, ok := resI.(Postable); ok && r.Method == "POST" {
			postable.DoPost(w, r)
		} else if deletable, ok := resI.(Deleteable); ok && r.Method == "DELETE" {
			deletable.DoDelete(w, r)
		} else {
			respond.NotAllowed(w)
		}
	}

}
