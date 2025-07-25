// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package applications

import (
	"slices"
)

type Collection struct {
	Apps      map[string]*DesktopApplication
	Mimetypes map[string]*Mimetype
}

func collect() Collection {
	var collection = Collection{}
	var defaultApps map[string][]string
	collection.Mimetypes = collectMimetypes()
	collection.Apps, defaultApps = collectApps()

	for mt, apps := range defaultApps {
		if mt, ok := collection.Mimetypes[mt]; ok {
			mt.Applications = apps
		}
	}

	for appId, app := range collection.Apps {
		for _, mimetypeId := range app.Mimetypes {
			if mimetype, ok := collection.Mimetypes[mimetypeId]; ok {
				mimetype.Applications = appendIfNotThere(mimetype.Applications, appId)
			}
		}
	}

	// Ensure that if, m1 is a subclass of m2, m1s applications contains all m2s applications

	var visited = make(map[string]bool)
	var getHandlersForSupertypes func(mt *Mimetype)
	getHandlersForSupertypes = func(mt *Mimetype) {
		if !visited[mt.Id] {
			for _, superId := range mt.SubClassOf {
				if superType, ok := collection.Mimetypes[superId]; ok {
					getHandlersForSupertypes(superType)
					for _, app := range superType.Applications {
						mt.Applications = appendIfNotThere(mt.Applications, app)
					}
				}
			}
			visited[mt.Id] = true
		}

	}

	for _, mt := range collection.Mimetypes {
		getHandlersForSupertypes(mt)
	}

	return collection
}

func appendIfNotThere(list []string, otherList ...string) []string {
	for _, other := range otherList {
		var found = false
		if slices.Contains(list, other) {
			found = true
		}
		if !found {
			list = append(list, other)
		}
	}
	return list
}

func remove(list []string, otherList ...string) []string {
	var pos = 0
	for i := 0; i < len(list); i++ {
		var found = false
		if slices.Contains(otherList, list[i]) {
			found = true
		}
		if !found {
			list[pos] = list[i]
			pos += 1
		}
	}
	return list[0:pos]
}
