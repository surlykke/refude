// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"time"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/action"
	"fmt"
)

const NotificationMediaType mediatype.MediaType = "application/vnd.org.refude.desktopnotification+json"

type Notification struct {
	resource.AbstractResource
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

func (n *Notification) getActions() []*action.Action {
	var actions = make([]*action.Action, 0, len(n.Actions))

	actions = append(actions,
		action.MakeAction(
			fmt.Sprintf("/actions/%d/a/dismiss", n.Id),
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
		actions = append(actions, action.MakeAction(
			fmt.Sprintf("/actions/%d/b/%s", n.Id, actionId),
			actionName,
			n.Subject,
			"",
			func() { conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".ActionInvoked", n.Id, actionIdCopy) },
		))
	}

	return actions
}
