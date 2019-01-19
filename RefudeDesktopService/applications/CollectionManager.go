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

func Run(updateStream chan<- resource.Update) {
	var collected = make(chan collection)

	go CollectAndWatch(collected)

	for applicationsAndMimetypes := range collected {
		var update = resource.Update{PrefixesToRemove: []resource.StandardizedPath{"/applications", "/mimetypes"}}

		for _, app := range applicationsAndMimetypes.applications {
			update.Mappings = append(update.Mappings, resource.Mapping{app.GetSelf(), app})
		}

		for _, mt := range applicationsAndMimetypes.mimetypes {
			update.Mappings = append(update.Mappings, resource.Mapping{mt.GetSelf(), mt})
			for _, alias := range mt.Aliases {
				var altPath = resource.Standardizef("/mimetypes/%s", alias)
				update.Mappings = append(update.Mappings, resource.Mapping{altPath, mt})
			}
		}

		updateStream <- update
	}
}

func reportError(msg string) {
	log.Println(msg)
}
