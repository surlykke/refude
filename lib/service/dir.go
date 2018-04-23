package service

import (
	"net/http"
	"github.com/surlykke/RefudeServices/lib/resource"
	"path"
	"strings"
	"github.com/surlykke/RefudeServices/lib/query"
	"fmt"
	"crypto/sha1"
	"sort"
)

/**
 * A standardized path is a path that is cleaned with path.Clean and has any leading '/' removed.
 * Eg: '//foo/baa' becomes 'foo/baa'
 *     '/foo/..//baa/.////muh' becomes 'foo/baa/muh'
 */
type StandardizedPath string

func Standardize(p string) StandardizedPath {
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


type SortedKeys []string

func (sk SortedKeys) Len() int {
	return len(sk)
}

func (sk SortedKeys) Less(i,j int) bool {
	return strings.Compare(sk[i], sk[j]) < 0
}

func (sk SortedKeys) Swap(i,j int) {
	sk[i],sk[j] = sk[j],sk[i]
}

type Dir struct {
	entries map[string]interface{}
	sortedKeys SortedKeys
	etag       string
}

func MakeEmptyDir() *Dir {
	return &Dir{make(map[string]interface{}), nil, ""}
}

func (d *Dir) updated() {
	d.sortedKeys = nil
	d.etag = ""
}

func (d *Dir) Map(sp StandardizedPath, res interface{}) {
	if first, remain := separate(sp); remain == "" {
		if _, ok := d.entries[first + "/"]; ok {
			panic("There's a directory here")
		}
		d.entries[first] = res
		d.updated()
	} else {
		if _,nonDirHere := d.entries[first[:len(first) - 1]]; nonDirHere {
			panic("There's a non-direcory here (" + first[:len(first) - 1])
		}
		subDir, dirHere := d.entries[first].(*Dir)
		if !dirHere {
			subDir = MakeEmptyDir()
			d.entries[first] = subDir
		}
		subDir.Map(remain, res)
	}
}

func (d *Dir) UnMap(sp StandardizedPath) {
	if first, remain:= separate(sp); remain == "" {
		if _,ok := d.entries[first] ; ok {
			delete(d.entries, string(first))
			d.updated()
		}
	} else {
		if subdir, ok := d.entries[first].(*Dir); ok {
			subdir.UnMap(remain)
		}
	}
}

func (d *Dir) MkDir(sp StandardizedPath) {
	if first, remain := separate(sp); remain == "" {
		if _,ok := d.entries[first]; ok {
			panic("There's a non-directory here")
		}

		if _,ok := d.entries[first + "/"]; !ok {
			d.entries[first + "/"] = MakeEmptyDir()
			d.updated()
		}
	} else {
		if _,ok := d.entries[first[:len(first) - 1]]; ok {
			panic("There's a non-directory here")
		}

		subdir, ok := d.entries[first].(*Dir)
		if !ok {
			subdir = MakeEmptyDir()
			d.entries[first] = subdir
		}
		subdir.MkDir(remain)
	}
}

func (d *Dir) Find(p StandardizedPath) (interface{}, bool) {
	if first, remain := separate(p); remain == "" {
		if res, ok := d.entries[first]; ok {
			return res, true
		} else if res, ok = d.entries[first + "/"]; ok {
			return res, true
		} else {
			return nil, false
		}
	} else if subdir, ok := d.entries[first].(*Dir); ok {
		return subdir.Find(remain)
	} else {
		return nil, false
	}
}

func (d *Dir) GET(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Query()) == 0 {
		if d.sortedKeys == nil {
			fmt.Println("sorting..")
			d.sortedKeys = make([]string, 0, len(d.entries))
			for key,_ := range d.entries {
				d.sortedKeys = append(d.sortedKeys, key)
			}
			sort.Sort(d.sortedKeys)
			h := sha1.New()
			for _,key := range d.sortedKeys {
				h.Write([]byte(key))
			}
			d.etag = fmt.Sprintf("\"%x\"", h.Sum(nil))
		}
		resource.JsonGET(d.sortedKeys, w)
	} else if len(r.URL.Query()) == 1 && len(r.URL.Query()["q"]) == 1 {
		if m, err := query.Parse(r.URL.Query()["q"][0]); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else {
			var matched = make([]interface{}, 0, len(d.entries))
			for _, res := range d.entries {
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


