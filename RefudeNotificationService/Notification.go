// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"time"
	"fmt"
	"github.com/surlykke/RefudeServices/lib"
)

const NotificationMediaType lib.MediaType = "application/vnd.org.refude.desktopnotification+json"

type Notification struct {
	lib.AbstractResource
	Id            uint32
	internalId    uint32
	Sender        string
	Subject       string
	Body          string
	Actions       map[string]string
	RelevanceHint int
	Expires       *time.Time `json:",omitempty"`
}

func (n *Notification) removeAfter(duration time.Duration) {
	time.AfterFunc(duration, func() { removals <- removal{n.Id, n.internalId, Expired} })
}

func (n *Notification) getActions() []*lib.Action {
	var actions = make([]*lib.Action, 0, len(n.Actions))

	actions = append(actions,
		lib.MakeAction(
			lib.Standardizef("/actions/%d/a/dismiss", n.Id),
			"Dismiss",
			n.Subject,
			"",
			func() {
				fmt.Println("Sending to removals..")
				removals <- removal{n.Id, 0, Dismissed}
				}))

	for actionId, actionName := range n.Actions {
		if actionId == "" {
			continue
		}
		var actionIdCopy = actionId
		actions = append(actions, lib.MakeAction(
			lib.Standardizef("/actions/%d/b/%s", n.Id, actionId),
			actionName,
			n.Subject,
			"",
			func() { conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".ActionInvoked", n.Id, actionIdCopy) },
		))
	}

	return actions
}
