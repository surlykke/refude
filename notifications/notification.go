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
	"sync"
	"time"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/watch"
)

type Urgency string

const (
	low              = "Low"
	normal           = "Normal"
	critical Urgency = "Critical"
)

type Notification struct {
	Id       uint32
	Sender   string
	Subject  string
	Body     string
	Created  time.Time
	Expires  time.Time `json:",omitempty"`
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
	return false
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

var flashResource *resource.Resource
var flashLock sync.Mutex

func getFlashResource() *resource.Resource {
	flashLock.Lock()
	defer flashLock.Unlock()
	return flashResource
}

func getFlash() *Notification {
	var res = getFlashResource()
	if res == nil {
		return nil
	} else {
		return res.Data.(*Notification)
	}
}

func setFlash(newFlashResource *resource.Resource) {
	flashLock.Lock()
	defer flashLock.Unlock()
	flashResource = newFlashResource
}

func somethingChanged() {
	watch.SomethingChanged("/notification/list")
	watch.DesktopSearchMayHaveChanged()
}
