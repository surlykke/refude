package main

import (
	"github.com/surlykke/RefudeServices/lib"
	"net/http"
)

const RegistryMediaType = "application/vnd.org.refude.actionregistry"

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
}

func getId() int {
	idLock.Lock()
	defer idLock.Unlock()
	i++
	return i
}
