// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package applications

import (
	"github.com/surlykke/RefudeServices/lib/resource"
	"log"
)

func Run(mappingsStream chan<- resource.Mappings) {
	var collected = make(chan collection)

	go CollectAndWatch(collected)

	for update := range collected {
		var mappings = resource.Mappings{
			PathsToRemove: []resource.StandardizedPath{},
			PrefixesToRemove: []resource.StandardizedPath{"/applications", "/mimetypes"},
			ResourcesToMap: make(map[resource.StandardizedPath]resource.Resource)}

		for _, app := range update.applications {
			mappings.ResourcesToMap[app.GetSelf()] = app
		}

		for _, mt := range update.mimetypes {
			mappings.ResourcesToMap[mt.GetSelf()] = mt
			for _, alias := range mt.Aliases {
				mappings.ResourcesToMap[resource.Standardizef("/mimetypes/%s", alias)] = mt
			}
		}

		mappingsStream <- mappings
	}
}

func reportError(msg string) {
	log.Println(msg)
}
