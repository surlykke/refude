// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"github.com/surlykke/RefudeServices/lib"
	"net/http"
	"time"
)

const RegistryMediaType = "application/vnd.org.refude.actionregistry+json"

type ActionRegistry struct {
	lib.AbstractResource
}

func MakeRegistry() *ActionRegistry {
	return &ActionRegistry{lib.AbstractResource{Self: "/registry", Mt: RegistryMediaType}}
}

func (ar *ActionRegistry) POST(w http.ResponseWriter, r *http.Request) {
	var registration = MakeNewRegistration()
	jm.Map(registration)
	w.Write([]byte(registration.Self))
	reap(registration.Self)
}

func reap(path lib.StandardizedPath) {
	var registration = jm.GetResource(path).GetRes().(*Registration)
	if registration != nil {
		if (registration.Expires.Before(time.Now())) {
			var actionPrefix = lib.Standardizef("/actions/%d", registration.id)
			jm.RemoveAll(registration.Self, actionPrefix)
		} else {
			time.AfterFunc(registration.Expires.Add(100*time.Millisecond).Sub(time.Now()), func() {
				reap(path)
			})
		}
	}
}

func getId() int {
	idLock.Lock()
	defer idLock.Unlock()
	i++
	return i
}
