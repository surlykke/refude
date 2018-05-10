package service

import (
	"net/http"
	"fmt"
	"sync"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/requestutils"
	"github.com/surlykke/RefudeServices/lib/query"
)

const LinksMediaType mediatype.MediaType = "application/vnd.org.refude.Links+json"

type Links struct {
	entries   map[string]mediatype.MediaType
	linksLock sync.Mutex
}

func MakeLinks() *Links {
	return &Links{entries: make(map[string]mediatype.MediaType)}
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
	fmt.Println("Into Links#Serve")
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
}

func (l *Links) getByteRepresentation(mediaType mediatype.MediaType) []byte {
	l.linksLock.Lock()
	defer l.linksLock.Unlock()
	if mediaType == "" {
		return mediatype.ToJSon(l.entries)
	} else {
		var matching = make(map[string]mediatype.MediaType, len(l.entries))
		for path, mt := range l.entries {
			if mediatype.MediaTypeMatch(mediaType, mt) {
				matching[path] = mt
			}
		}
		return mediatype.ToJSon(matching)
	}
}
