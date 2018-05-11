// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"github.com/surlykke/RefudeServices/lib/service"
	"fmt"
	"time"
	"net/http"
	"github.com/surlykke/RefudeServices/lib/requestutils"
	"github.com/surlykke/RefudeServices/lib/resource"
)

var removals = make(chan removal)


func Run() {
	var updates = make(chan *Notification)
	var removals = make(chan removal)
	var reap = make(chan struct{})

	var pendingTimouts = make(map[uint32]time.Time)

	go DoDBus(updates, removals)

	for {
		select {
		case notification := <-updates:
			service.Map(path(notification.Id), resource.MakeJsonResource(*notification, NotificationMediaType))
			if notification.Expires != nil {
				fmt.Println("do afterFunc")
				pendingTimouts[notification.Id] = *notification.Expires
				time.AfterFunc(notification.Expires.Sub(time.Now()), func(){fmt.Println("Signal reaper"); reap <- struct{}{}})
			}
		case rem := <-removals:

			if res, ok := service.Unmap(path(rem.id)); ok {
				if res.Mt() == NotificationMediaType {
					notificationClosed(rem.id, rem.reason)
				}
			}
		case _ = <-reap:
			fmt.Println("Reaping..")
			for id,expires := range pendingTimouts {
				fmt.Println("compare", expires, "to", time.Now())
				if expires.Before(time.Now()) {
					fmt.Println("deleting")
					delete(pendingTimouts, id)
					if _, ok := service.Unmap(path(id)); ok {
						notificationClosed(id, Expired)
					}
				}
			}
		}
	}
}

func path(id uint32) string {
	return fmt.Sprintf("/notifications/%d", id)
}

func reaper(reap chan struct{}) {
	for {
		time.Sleep(time.Second)
		reap <- struct{}{}
	}
}

func (n *Notification) POST(w http.ResponseWriter, r *http.Request) {
		action := requestutils.GetSingleQueryParameter(r, "action", "default")
		conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".ActionInvoked", n.Id, action)
		w.WriteHeader(http.StatusAccepted)
}

func (n *Notification) DELETE(w http.ResponseWriter, r *http.Request) {
	removals <- removal{n.Id, Dismissed}
	w.WriteHeader(http.StatusAccepted)
}

