// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"github.com/surlykke/RefudeServices/lib/resource"
	"log"
)

type launchEvent struct {
	id   string
	time int64
}

func Run() {
	var collected = make(chan collection)

	go CollectAndWatch(collected)

	for {
		select {
		case update := <-collected:
			resourceHandler.RemoveAll("/applications")
			resourceHandler.RemoveAll("/actions")
			for _, app := range update.applications {
				resourceHandler.Map(app)
			}
			resourceHandler.RemoveAll("/mimetypes")
			for _, mt := range update.mimetypes {
				resourceHandler.Map(mt)
				for _,alias := range mt.Aliases {
					resourceHandler.MapTo(resource.Standardizef("/mimetypes/%s", alias), mt)
				}
			}
		}
	}
}

func reportError(msg string) {
	log.Println(msg)
}


