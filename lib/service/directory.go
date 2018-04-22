package service

import (
	"net/http"
	"github.com/surlykke/RefudeServices/lib/resource"
	"path"
	"strings"
	"github.com/surlykke/RefudeServices/lib/query"
	"fmt"
)

/**
 * A standardized path is a path that is cleaned with path.Clean and has any leading '/' removed.
 * Eg: '//foo/baa' becomes 'foo/baa'
 *     '/foo/..//baa/.////muh' becomes 'foo/baa/muh'
 */
type StandardizedPath string

func MakeStandardizedPath(p string) StandardizedPath {
	var cleanPath = path.Clean(p)
	if len(cleanPath) > 0 && cleanPath[0] == '/' {
		return StandardizedPath(cleanPath[1:])
	} else {
		return StandardizedPath(cleanPath)
	}
}

func separate(sp StandardizedPath) (string, StandardizedPath) {
	if slashPos := strings.IndexByte(string(sp), '/'); slashPos > -1 {
		return string(sp[0:slashPos + 1]), sp[slashPos + 1:]
	} else {
		return string(sp), ""
	}
}


type Dir map[string]interface{}

func (d Dir) Map(sp StandardizedPath, res interface{}) {
	if first, remain := separate(sp); remain == "" {
		if _, ok := d[first + "/"]; ok {
			panic("There's a directory here")
		}
		d[first] = res
	} else {
		if _,nonDirHere := d[first[:len(first) - 1]]; nonDirHere {
			panic("There's a non-direcory here (" + first[:len(first) - 1])
		}
		subDir, dirHere := d[first].(Dir)
		if !dirHere {
			subDir = make(Dir)
			d[first] = subDir
		}
		subDir.Map(remain, res)
	}
}

func (d Dir) UnMap(sp StandardizedPath) {
	if first, remain:= separate(sp); remain == "" {
		delete(d, string(first))
	} else {
		if subdir, ok := d[first].(Dir); ok {
			subdir.UnMap(remain)
			if len(subdir) == 0 {
				delete(d, first)
			}
		}
	}
}

func (d Dir) Find(p StandardizedPath) (interface{}, bool) {
	if first, remain := separate(p); remain == "" {
		if res, ok := d[first]; ok {
			return res, true
		} else if res, ok = d[first + "/"]; ok {
			return res, true
		} else {
			return nil, false
		}
	} else if subdir, ok := d[first].(Dir); ok {
		return subdir.Find(remain)
	} else {
		return nil, false
	}
}

func (d Dir) GET(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Query()) == 0 {
		paths := make([]string, 0, len(d))
		for key, _ := range d {
			paths = append(paths, key)
		}
		resource.JsonGET(paths, w)
		return
	} else if len(r.URL.Query()) == 1 && len(r.URL.Query()["q"]) == 1 {
		if m, err := query.Parse(r.URL.Query()["q"][0]); err != nil {
			fmt.Println("Parsing error: ", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else {
			var matched = make([]interface{}, 0, len(d))
			for _, res := range d {
				if m(res) {
					matched = append(matched, res)
				}
			}
			resource.JsonGET(matched, w)
		}
	} else {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}

/*
	If sp contains a "/", return start of sp up til first "/" + from just after "/" to end
    else return sp
    Eg: "foo/baa/muh" -> "foo", "baa/muh"
        "foo" -> "foo", ""
 */


