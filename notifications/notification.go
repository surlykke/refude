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
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
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
	resource.BaseResource
	NotificationId uint32
	Sender         string
	Created        int64
	Expires        int64
	Urgency        Urgency
	NActions       map[string]string `json:"actions"`
	Hints          map[string]interface{}
	iconName       string
	IconSize       uint32 `json:",omitempty"`
}



func (n *Notification) Actions() link.ActionList {
	var ll = link.ActionList{}
	if actionDesc, ok := n.NActions["default"]; ok {
		ll = append(ll, link.MkAction("default", actionDesc, ""))
	}
	for actionId, actionDesc := range n.NActions {
		if actionId != "default" {
			ll = append(ll, link.MkAction(actionId, actionDesc, ""))
		}
	}
	// FIXME ll = append(ll, link.Make(self, "Dismiss", "", relation.Delete))

	return ll
}


func (n *Notification) DeleteAction() (string, bool) {
	return "Dismiss", true
}

func (n *Notification) RelevantForSearch() bool {
	return n.Urgency == Critical || (len(n.NActions) > 0 && n.Urgency == Normal && n.Created+60000 > time.Now().UnixMilli())
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

var Notifications = resource.MakeCollection[*Notification]()




