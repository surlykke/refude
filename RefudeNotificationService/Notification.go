// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"github.com/surlykke/RefudeServices/lib"
	"time"
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

/**
 * If the notification has a default action we build an action for that
 * We always build a dissmiss action
 */
func (n *Notification) buildActions() []*lib.Action {
	var res = make([]*lib.Action, 0, 2)

	if _, ok := n.Actions["default"]; ok {
		res = append(res, lib.MakeAction(
			lib.Standardizef("/actions/%d", n.Id),
			n.Actions["default"],
			n.Subject,
			"", // FIXME
			func() {
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".ActionInvoked", n.Id, "default")
			},
		))
	}
	res = append(res, lib.MakeAction(
		lib.Standardizef("/actions/%d", n.Id),
		n.Subject,
		"Dismiss",  // TODO i18n
		"", // TODO
		func() {
			removals <- removal{n.Id, 0, Dismissed}
		},
	))

	return res
}
