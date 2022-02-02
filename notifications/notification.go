// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"net/http"
	"time"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/watch"
)

type Urgency uint8

const (
	Low      Urgency = 0
	Normal           = 1
	Critical         = 2
)

const flashTimeoutLow time.Duration = 3 * time.Second
const flashTimeoutNormal time.Duration = 8 * time.Second
const _50ms = 50 * time.Millisecond

type Notification struct {
	Id       uint32
	Sender   string
	Subject  string
	Body     string
	Created  time.Time
	Expires  time.Time
	Urgency  Urgency
	NActions map[string]string `json:"actions"`
	Hints    map[string]interface{}
	iconName string
	IconSize uint32 `json:",omitempty"`
}

func (n *Notification) GetPostActions() []resource.Action {
	var actions []resource.Action

	if actionDesc, ok := n.NActions["default"]; ok {
		actions = []resource.Action{{Title: actionDesc}}
	}
	for actionId, actionDesc := range n.NActions {
		if actionId != "default" {
			actions = append(actions, resource.Action{Id: actionId, Title: actionDesc})
		}
	}

	return actions
}

func (n *Notification) GetDeleteAction() *resource.Action {
	return &resource.Action{Title: "Dismiss", Icon: ""}
}

func (n *Notification) DoPost(w http.ResponseWriter, r *http.Request) {

	var action = requests.GetSingleQueryParameter(r, "action", "default")

	if _, ok := n.NActions[action]; ok {
		if err := conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".ActionInvoked", n.Id, action); err != nil {
			respond.ServerError(w, err)
		} else {
			respond.Accepted(w)
		}
	} else {
		respond.NotFound(w)
	}
}

func (n *Notification) DoDelete(w http.ResponseWriter, r *http.Request) {
	removeNotification(n.Id, Dismissed)
	respond.Accepted(w)
}

var Notifications = resource.MakeList("/notification/list")

func getFlashResource() *resource.Resource {
	var found *resource.Resource

	Notifications.Walk(func(res *resource.Resource) {
		var n = res.Data.(*Notification)
		if found == nil || found.Data.(*Notification).Urgency < n.Urgency {
			if n.Urgency == Critical ||
				n.Urgency == Normal && n.Created.After(time.Now().Add(-flashTimeoutNormal)) ||
				n.Urgency == Low && n.Created.After(time.Now().Add(-flashTimeoutLow)) {
				found = res
			}
		}
	})
	return found
}

func somethingChanged() {
	watch.SomethingChanged("/notification/flash")
	watch.SomethingChanged("/notification/list")
	watch.DesktopSearchMayHaveChanged()
}
