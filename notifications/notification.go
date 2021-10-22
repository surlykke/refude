// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"fmt"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/watch"
)

type Urgency string

const (
	critical Urgency = "Critical"
	normal           = "Normal"
	low              = "Low"
)

type Notification struct {
	Id       uint32
	Sender   string
	Subject  string
	Body     string
	Created  uint64
	Expires  uint64 `json:",omitempty"`
	Urgency  Urgency
	Actions  map[string]string
	Hints    map[string]interface{}
	iconName string
}

func (n *Notification) Links(path string) link.List {
	var ll = make(link.List, 0, 3)
	ll = ll.Add(path, "Dismiss", "", relation.Delete)

	for actionId, actionDesc := range n.Actions {
		if actionId == "default" {
			ll = ll.Add(path, actionDesc, "", relation.DefaultAction)
		} else {
			ll = ll.Add(path+"?action="+actionId, actionDesc, "", relation.DefaultAction)
		}
	}

	return ll
}

func (n *Notification) ForDisplay() bool {
	return n.Urgency == critical ||
		n.Created > nowMillis()-6000 ||
		len(n.Actions) > 0
}

func (n *Notification) DoPost(w http.ResponseWriter, r *http.Request) {
	fmt.Println("notification doPost")
	var action = requests.GetSingleQueryParameter(r, "action", "default")
	fmt.Println("Action:", action)
	if _, ok := n.Actions[action]; ok {
		fmt.Println("Emitting")
		if err := conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".ActionInvoked", n.Id, action); err != nil {
			fmt.Println("Got error", err)
			respond.ServerError(w, err)
		} else {
			fmt.Println("ok")
			respond.Accepted(w)
		}
	} else {
		fmt.Println("not found")
		respond.NotFound(w)
	}
}

func (n *Notification) DoDelete(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Deleting ", n.Id)
	removals <- removal{n.Id, Dismissed}
	respond.Accepted(w)
}

var Notifications = resource.MakeRevertedList("/notification/list")

// Notifiation collection

func removeNotification(id uint32, reason uint32) {
	fmt.Println("In removeNotification")
	var path = fmt.Sprintf("/notification/%X", id)
	if found := Notifications.Delete(path); found {
		conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, reason)
		fmt.Println("somethingChanged...")
		somethingChanged()
	}
}

func removeExpired() {
	fmt.Println("removeExpired")
	var somethingExpired = false
	for _, res := range Notifications.GetAll() {
		var notification = res.Data.(*Notification)
		if notification.Expires > 0 && notification.Expires < nowMillis() {
			Notifications.Delete(fmt.Sprintf("/notification/%X", notification.Id))
			conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", notification.Id, Expired)
		}
	}
	if somethingExpired {
		somethingChanged()
	}
}

func somethingChanged() {
	watch.SomethingChanged("/notification/list")
	watch.DesktopSearchMayHaveChanged()
}
