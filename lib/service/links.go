package service

import (
	"github.com/surlykke/RefudeServices/lib/resource"
	"net/http"
	"fmt"
)


const LinksMediaType resource.MediaType = "application/vnd.org.refude.links+json"

type linkEntry struct {
	path      string
	mediaType resource.MediaType
}

type links struct {
	resource.ByteResource
	Entries []linkEntry
}

func makeLinks(entries []linkEntry) *links {
	return &links{ByteResource: resource.MakeByteResource(LinksMediaType), Entries: entries}
}

func (l *links) addEntry(path string, mediaType resource.MediaType) *links {
	for _, entry := range l.Entries {
		if path == entry.path {
			return l
		}
	}

	var newEntries = make([]linkEntry, len(l.Entries) + 1, len(l.Entries) + 1)
	copy(newEntries, l.Entries)
	newEntries[len(newEntries) - 1] = linkEntry{path, mediaType}
	return makeLinks(newEntries)
}

func (l *links) removeEntry(path string) *links {
	for i := 0; i < len(l.Entries); i++ {
		if l.Entries[i].path == path {
			var newEntries= make([]linkEntry, len(l.Entries)-1, len(l.Entries)-1)
			copy(newEntries, l.Entries[0:i])
			copy(newEntries[i:], l.Entries[i+1:])
			return makeLinks(newEntries)
		}
	}
	return l
}

func (l *links) Update() resource.Resource {
	var m = make(map[string]resource.MediaType)
	for _, entry := range l.Entries {
		m[entry.path] = entry.mediaType
	}
	var copy = makeLinks(l.Entries)
	copy.SetBytes(resource.ToJSon(m))
	return copy
}

func (l *links) GET(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Query()) == 0 {
		l.ByteResource.GET(w, r)
	} else if flatParams, err := resource.GetSingleParams(r, "type"); err != nil {
		fmt.Println("Error links.GET: ", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		var mediaType = resource.MediaType(flatParams["type"])
		var filteredEntries = make(map[string]resource.MediaType, len(l.Entries))
		for _, entry := range l.Entries {
			if mediaType == "" || mediaType == entry.mediaType {
				filteredEntries[entry.path] = entry.mediaType
			}
		}
		w.Write(resource.ToJSon(filteredEntries))
	}
}

