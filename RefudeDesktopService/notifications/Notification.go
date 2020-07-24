// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"net/http"
	"time"

	"github.com/surlykke/RefudeServices/lib/respond"
)

type Notification struct {
	self     string
	Id       uint32
	Sender   string
	Subject  string
	Body     string
	IconName string `json:",omitempty"`
	Created  time.Time
	Expires  time.Time `json:",omitempty"`
	Actions  map[string]string
	Hints    map[string]interface{}
}

func (n *Notification) ToStandardFormat() *respond.StandardFormat {
	return &respond.StandardFormat{
		Self:     n.self,
		Type:     "notification",
		Title:    n.Subject,
		Comment:  n.Body,
		IconName: n.IconName,
		OnPost:   n.Actions["default"],
		OnDelete: "Dismiss",
		Data:     n,
	}
}

func (n *Notification) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, n.ToStandardFormat())
	} else if r.Method == "POST" {
		// TODO otheractions
		if n.haveDefaultAction() {
			respond.AcceptedAndThen(w, func() {
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".ActionInvoked", n.Id, "default")
			})
		} else {
			respond.NotAllowed(w)
		}
	} else if r.Method == "DELETE" {
		respond.AcceptedAndThen(w, func() { removals <- removal{id: n.Id, reason: Dismissed} })

	} else {
		respond.NotAllowed(w)
	}
}

func (n *Notification) haveDefaultAction() bool {
	_, ok := n.Actions["default"]
	return ok
}
