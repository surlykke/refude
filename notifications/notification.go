// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package notifications

import (
	"net/http"
	"time"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

type Urgency uint8

const (
	Low      Urgency = 0
	Normal           = 1
	Critical         = 2
)

const flashTimeoutLow time.Duration = 2 * time.Second
const flashTimeoutNormal time.Duration = 6 * time.Second
const _50ms = 50 * time.Millisecond

type Notification struct {
	NotificationId uint32
	Sender         string
	Subject        string
	Body           string
	Created        int64
	Expires        int64
	Urgency        Urgency
	NActions       map[string]string `json:"actions"`
	Hints          map[string]interface{}
	iconName       string
	IconSize       uint32 `json:",omitempty"`
}

func (n *Notification) Id() uint32 {
	return n.NotificationId
}

func (n *Notification) Presentation() (title string, comment string, icon link.Href, profile string) {
	return n.Subject, n.Body, link.IconUrl(n.iconName), "notification"
}

func (n *Notification) Links(self, term string) link.List {
	var ll = link.List{}
	if actionDesc, ok := n.NActions["default"]; ok {
		if searchutils.Match(term, actionDesc) > -1 {
			ll = append(ll, link.Make(self+"?action=default", actionDesc, "", relation.DefaultAction))
		}
	}
	for actionId, actionDesc := range n.NActions {
		if searchutils.Match(term, actionDesc) > -1 {
			if actionId != "default" {
				ll = append(ll, link.Make(self+"?action="+actionId, actionDesc, "", relation.Action))
			}
		}
	}
	ll = append(ll, link.Make(self, "Dismiss", "", relation.Delete))

	return ll
}

func (n *Notification) DoPost(w http.ResponseWriter, r *http.Request) {

	var action = requests.GetSingleQueryParameter(r, "action", "default")

	if _, ok := n.NActions[action]; ok {
		if err := conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".ActionInvoked", n.NotificationId, action); err != nil {
			respond.ServerError(w, err)
		} else {
			respond.Accepted(w)
		}
	} else {
		respond.NotFound(w)
	}
}

func (n *Notification) DoDelete(w http.ResponseWriter, r *http.Request) {
	removeNotification(n.NotificationId, Dismissed)
	respond.Accepted(w)
}

var Notifications = resource.MakePublishingCollection[uint32, *Notification]("/notification/", "/search")


func Search(term string) link.List {
	var rank = func(n *Notification) int {
		if n.Urgency == Critical || (len(n.NActions) > 0 && n.Urgency == Normal && n.Created+60000 > time.Now().UnixMilli()) {
			return searchutils.Match(term, n.Subject)
		} else {
			return -1
		}
	}
	return Notifications.ExtractLinks(rank) 
}



