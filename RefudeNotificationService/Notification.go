// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"net/http"
	"github.com/surlykke/RefudeServices/lib/resource"
)

type Notification struct {
	Id            uint32
	Sender        string
	Subject       string
	Body          string
	Actions       map[string]string
	RelevanceHint int
	eTag          string
}

func (n *Notification) GET(w http.ResponseWriter, r *http.Request) {
	resource.JsonGET(n, w)
}

func (n *Notification) POST(w http.ResponseWriter, r *http.Request) {
	action := resource.GetSingleQueryParameter(r, "action", "default")
	conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE + ".ActionInvoked", n.Id, action)
	w.WriteHeader(http.StatusAccepted)
}

func (n *Notification) ETag() string {
	return n.eTag
}

func (n *Notification) DELETE(w http.ResponseWriter, r *http.Request) {
	close(r.URL.Path, "", Dismissed)
}
