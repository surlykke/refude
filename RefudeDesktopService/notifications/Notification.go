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

	"github.com/surlykke/RefudeServices/lib/respond"
)

type Notification struct {
	Id        uint32
	Sender    string
	Subject   string
	Body      string
	IconName  string `json:",omitempty"`
	Image     string `json:",omitempty"`
	imagePath string
	Created   time.Time
	Expires   time.Time `json:",omitempty"`
	Actions   map[string]string
}

func (n *Notification) ToStandardFormat() *respond.StandardFormat {
	return &respond.StandardFormat{
		Self:     notificationSelf(n.Id),
		Type:     "notification",
		Title:    n.Subject,
		Comment:  n.Body,
		IconName: n.IconName,
		OnPost:   n.Actions["default"],
		Data:     n,
	}
}

func (n *Notification) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, n.ToStandardFormat())
	} else if r.Method == "POST" && n.haveDefaultAction() {
		conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".ActionInvoked", n.Id, "default")
	} else {
		respond.NotAllowed(w)
	}
}

func (n *Notification) haveDefaultAction() bool {
	_, ok := n.Actions["default"]
	return ok
}

type NotificationImage struct {
	imagePath string
}

func (ni *NotificationImage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, ni.imagePath)
	} else {
		respond.NotAllowed(w)
	}
}

func notificationSelf(id uint32) string {
	return fmt.Sprintf("/notification/%d", id)
}

func notificationImageSelf(id uint32) string {
	return fmt.Sprintf("/notificationimage/%d", id)
}
