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

func collect() (map[string]*DesktopApplication, map[string]*Mimetype) {
	var apps, defaultApps = collectApps() 
	var mimetypes =  collectMimetypes()

	for mimetypeId, appIds := range defaultApps {
		if mimetype, ok := mimetypes[mimetypeId]; ok {
			mimetype.Applications = appIds
		}
	}

	for appId, app := range apps {
		for _, mimetypeId := range app.Mimetypes {
			if mimetype, ok := mimetypes[mimetypeId]; ok {
				mimetype.Applications = appendIfNotThere(mimetype.Applications, appId)
			}
		}
	}

	// Ensure that if, m1 is a subclass of m2, m1s applications contains all m2s applications
	// FIXME: Wrong?
	var visited = make(map[string]bool)
	var getHandlersForSupertypes func(mt *Mimetype)
	getHandlersForSupertypes = func(mt *Mimetype) {
		if !visited[mt.Id] {
			for _, superId := range mt.SubClassOf {
				if superType, ok := mimetypes[superId]; ok {
					getHandlersForSupertypes(superType)
					for _, app := range superType.Applications {
						mt.Applications = appendIfNotThere(mt.Applications, app)
					}
				}
			}
			visited[mt.Id] = true
		}

	}

	for _, mt := range mimetypes {
		getHandlersForSupertypes(mt)
	}

	return apps, mimetypes
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
	for i := range list {
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
