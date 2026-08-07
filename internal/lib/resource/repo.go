// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package resource

import (
	"net/http"
	"strings"
	"sync"

	"github.com/surlykke/refude/internal/lib/respond"
	"github.com/surlykke/refude/internal/lib/utils"
)

var m = make(map[string]Servable)
var lock sync.Mutex

func get(path string) (Servable, bool) {
	lock.Lock()
	defer lock.Unlock()
	v, ok := m[path]
	return v, ok
}

func Get[V Servable](path string) (V, bool) {
	if s, ok := get(path); ok {
		if v, ok := s.(V); ok {
			return v, true
		}
	}
	var v V
	return v, false
}

func getByPrefix(prefix string) []Servable {
	lock.Lock()
	defer lock.Unlock()
	var l = make([]Servable, 0, 100)
	for p, s := range m {
		if strings.HasPrefix(p, prefix) {
			l = append(l, s)
		}
	}
	return l
}

func GetByPrefix[V Servable](prefix string) []V {
	var servables = getByPrefix(prefix)
	var list = make([]V, 0, len(servables))
	for _, s := range servables {
		if v, ok := s.(V); ok {
			list = append(list, v)
		}
	}

	return list
}

func Put(s Servable) {
	lock.Lock()
	defer lock.Unlock()
	m[s.GetBase().Path] = s
}

func Remove(path string) (Servable, bool) {
	lock.Lock()
	defer lock.Unlock()
	s, ok := m[path]
	if ok {
		delete(m, path)
	}
	return s, ok
}

func RemoveByPrefix(prefix string) {
	lock.Lock()
	defer lock.Unlock()
	for p, _ := range m {
		if strings.HasPrefix(p, prefix) {
			delete(m, p)
		}
	}
}

func Replace[V Servable](newVals []V, prefix string) {
	lock.Lock()
	defer lock.Unlock()
	for p, _ := range m {
		if strings.HasPrefix(p, prefix) {
			delete(m, p)
		}
	}
	for _, v := range newVals {
		m[v.GetBase().Path] = v
	}
}

func GetForSearch(prefixes ...string) []Base {
	var bases = make([]Base, 0, len(m))
	lock.Lock()
	defer lock.Unlock()

	for p, s := range m {
		for _, prefix := range prefixes {
			if strings.HasPrefix(p, prefix) && !s.OmitFromSearch() {
				bases = append(bases, *s.GetBase())
			}
		}
	}
	return bases
}

func GetPaths() []string {
	var paths = make([]string, 0, len(m))
	lock.Lock()
	defer lock.Unlock()
	for p, _ := range m {
		paths = append(paths, p)
	}
	return paths
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/") {
		if "GET" == r.Method {
			respond.AsJson(w, getByPrefix(r.URL.Path))
		} else {
			respond.NotAllowed(w)
		}
	} else {
		s, ok := get(r.URL.Path)
		if !ok {
			respond.NotFound(w)
			return
		}
		if "GET" == r.Method {
			respond.AsJson(w, s)
		} else if postable, ok := s.(Postable); ok && "POST" == r.Method {
			if ok, err := postable.DoPost(utils.QueryParam(r, "action")); err != nil {
				respond.ServerError(w, err)
			} else if !ok {
				respond.NotFound(w)
			} else {
				respond.Accepted(w)
			}
		} else if deleteable, ok := s.(Deleteable); ok && "DELETE" == r.Method {
			if err := deleteable.DoDelete(); err != nil {
				respond.ServerError(w, err)
			} else {
				respond.Accepted(w)
			}
		} else {
			respond.NotAllowed(w)
		}
	}
}
