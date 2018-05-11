// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package service

import (
	"net/http"
	"sync"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/requestutils"
	"github.com/surlykke/RefudeServices/lib/query"
	"time"
	"fmt"
)

const LinksMediaType mediatype.MediaType = "application/vnd.org.refude.Links+json"

type Links struct {
	entries   map[string]mediatype.MediaType
	linksLock sync.Mutex
	cache     map[mediatype.MediaType][]byte
}

func MakeLinks() *Links {
	return &Links{entries: make(map[string]mediatype.MediaType), cache: make(map[mediatype.MediaType][]byte)}
}

// Caller must have linksLock
func (l *Links) clearCache() {
	if len(l.cache) > 0 {
		l.cache = make(map[mediatype.MediaType][]byte)
	}
}

func (l *Links) Mt() mediatype.MediaType {
	return LinksMediaType
}

func (l *Links) Match(m query.Matcher) bool {
	return false
}

func (l *Links) addLinkEntry(path standardizedPath, mediaType mediatype.MediaType) {
	l.linksLock.Lock()
	defer l.linksLock.Unlock()
	l.entries[string(path[1:])] = mediaType
}

func (l *Links) removeLinkEntry(path standardizedPath) {
	l.linksLock.Lock()
	defer l.linksLock.Unlock()
	delete(l.entries, string(path)[1:])
}

// ----------------------------------------------------------------------------

func (l *Links) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var t1 = time.Now()
	if r.Method == "GET" {
		if flattenedParameterMap, err := requestutils.GetSingleParams(w, r, "type"); err != nil {
			requestutils.ReportUnprocessableEntity(w, err)
		} else {
			var mediaType= mediatype.MediaType(flattenedParameterMap["type"])
			w.Header().Set("Content-Type", string(LinksMediaType))
			w.Write(l.getByteRepresentation(mediaType))
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	var t2 = time.Now()
	fmt.Println("links.ServeHTTP took", t2.Sub(t1).Nanoseconds())
}

func (l *Links) getByteRepresentation(mediaType mediatype.MediaType) []byte {
	l.linksLock.Lock()
	defer l.linksLock.Unlock()
	data, ok := l.cache[mediaType]
	if !ok {
		if mediaType == "" {
			data = mediatype.ToJSon(l.entries)
		} else {
			var matching= make(map[string]mediatype.MediaType, len(l.entries))
			for path, mt := range l.entries {
				if mediatype.MediaTypeMatch(mediaType, mt) {
					matching[path] = mt
				}
			}
			data = mediatype.ToJSon(matching)
		}
		l.cache[mediaType] = data
	}
	return data
}
