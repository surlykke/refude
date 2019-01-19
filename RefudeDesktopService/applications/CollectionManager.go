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

func Run(resourceMap *resource.JsonResourceMap) {
	var collected = make(chan collection)

	go CollectAndWatch(collected)

	for applicationsAndMimetypes := range collected {
		var mappings = []resource.Mapping{}

		for _, app := range applicationsAndMimetypes.applications {
			mappings = append(mappings, resource.Mapping{app.GetSelf(), app})
		}

		for _, mt := range applicationsAndMimetypes.mimetypes {
			mappings = append(mappings, resource.Mapping{mt.GetSelf(), mt})
			for _, alias := range mt.Aliases {
				var altPath = resource.Standardizef("/mimetypes/%s", alias)
				mappings = append(mappings, resource.Mapping{altPath, mt})
			}
		}

		resourceMap.Update([]resource.StandardizedPath{"/mimetypes", "/applications"}, mappings)
	}
}

func reportError(msg string) {
	log.Println(msg)
}
