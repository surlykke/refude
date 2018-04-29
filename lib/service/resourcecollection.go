package service

import (
	"net/http"
	"github.com/surlykke/RefudeServices/lib/resource"
	"strings"
	"fmt"
	"sync"
)

// A standardized path is a path that starts with '/' and has no double slashes
type standardizedPath string

/** transform a path to a standardized path
 * Watered down version of path.Clean. Replace any sequence of '/' with single '/'
 * Remove ending '/'
 * We do not resolve '..', (so '/foo/../baa' is different from '/baa')
 * Examples:
 *       '//foo/baa' becomes '/foo/baa'
 *       '/foo///baa/////muh/' becomes '/foo/baa/muh'
 *       '/foo/..//baa//' becomes '/foo/../baa'
 */
func standardize(p string) standardizedPath {
	if len(p) == 0 || p[0] != '/' {
		panic(fmt.Sprintf("path must start with '/': '%s'", p))
	}

	var buffer, pos, justSawSlash = make([]byte, len(p), len(p)), 0, false
	for i := 0; i < len(p); i++ {
		if !justSawSlash || p[i] != '/' {
			buffer[pos] = p[i]
			pos++
		}
		justSawSlash = p[i] == '/'
	}

	if buffer[pos-1] == '/' {
		return standardizedPath(buffer[:pos-1])
	} else {
		return standardizedPath(buffer[:pos])
	}

}

/**
	Break standardized path into dir-part and base-part
    '/foo/baa/res' -> '/foo/baa', 'res'
    '/foo/baa' -> '/foo', 'baa'
 */
func separate(sp standardizedPath) (standardizedPath, string) {
	if len(sp) == 0 {
		panic("Separating empty string")
	} else {
		var pos = strings.LastIndexByte(string(sp[:len(sp)-1]), '/')
		return sp[:pos], string(sp[pos+1:])
	}
}


var mutex sync.Mutex
var rc = make(map[standardizedPath]resource.Resource)

func init() {
	rc["/links"] = makeLinks(nil)
	put("/search", &Search{})
}

// TODO: Note about threadsafety

func put(sp standardizedPath, res resource.Resource) {
	if sp == "/links" {
		panic("Attempt to map on \"/links\"")
	}

	rc[sp] = res
	var l = rc["/links"].(*links)
	rc["/links"] = l.addEntry(string(sp[1:]), res.MediaType())
}

func unput (sp standardizedPath) {
	if sp == "/links" {
		panic("Attempt to unmap \"/links\"")
	}

	delete(rc, sp)
	var l = rc["/links"].(*links)
	rc["/links"] = l.removeEntry(string(sp[1:]))
}

func findForServing(path standardizedPath) (resource.Resource, bool) {
	mutex.Lock()
	defer mutex.Unlock()
	var res resource.Resource
	var ok bool
	if res, ok = rc[path]; ok {
		if res2 := res.Update(); res2 != nil {
			rc[path] = res2
			res = res2
		}
		return res, true
	}
	return nil, false
}


// ------------------------------------ Public ----------------------------------------------------

func RemoveAll(dirpath string) {
	var lookFor = string(standardize(dirpath) + "/")
	mutex.Lock()
	defer mutex.Unlock()
	for path, _ := range rc {
		if strings.HasPrefix(string(path), lookFor) {
			unput(path)
		}
	}
}

func MapAll(newEntries map[string]resource.Resource) {
	mutex.Lock()
	defer mutex.Unlock()
	for path, res := range newEntries {
		sp := standardize(path)
		put(sp, res)
	}

}

func Map(path string, res resource.Resource) {
	sp := standardize(path)
	mutex.Lock()
	defer mutex.Unlock()
	put(sp, res)
}

func Unmap(path string) {
	sp := standardize(path)
	mutex.Lock()
	defer mutex.Unlock()
	unput(sp)
}

func UnMapIfMatch(path string, eTag string) bool {
	sp := standardize(path)
	mutex.Lock()
	defer mutex.Unlock()
	if res, ok := rc[sp]; ok {
		if res.ETag() == eTag {
			unput(sp)
			return true
		}
	}
	return false;
}

func Has(path string) bool {
	sp := standardize(path)
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := rc[sp]; ok {
		return true
	} else {
		return false
	}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sp := standardize(r.URL.Path)
	if res, ok := findForServing(sp); ok {
		switch r.Method {
		case "GET": res.GET(w, r)
		case "POST": res.POST(w, r)
		case "PATCH": res.PATCH(w, r)
		case "DELETE": res.DELETE(w, r)
		default: w.WriteHeader(http.StatusMethodNotAllowed)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
