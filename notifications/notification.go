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
	"time"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/watch"
)

type Urgency uint8

const (
	low      Urgency = 0
	normal           = 1
	critical         = 2
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
	Actions  map[string]string
	Hints    map[string]interface{}
	iconName string
	IconSize uint32 `json:",omitempty"`
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
	return n.Urgency == critical || len(n.Actions) > 0
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
	removeNotification(n.Id, Dismissed)
	respond.Accepted(w)
}

var Notifications = resource.MakeList("/notification/list")

func getFlashResource() *resource.Resource {
	var found *resource.Resource

	Notifications.Walk(func(res *resource.Resource) {
		var n = res.Data.(*Notification)
		if found == nil || found.Data.(*Notification).Urgency < n.Urgency {
			if n.Urgency == critical ||
				n.Urgency == normal && n.Created.After(time.Now().Add(-flashTimeoutNormal)) ||
				n.Urgency == low && n.Created.After(time.Now().Add(-flashTimeoutLow)) {
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
