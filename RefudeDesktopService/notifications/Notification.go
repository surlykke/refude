// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"fmt"
	"time"

	"github.com/surlykke/RefudeServices/lib/respond"
)

type Notification struct {
	Id       uint32
	Sender   string
	Subject  string
	Body     string
	IconName string `json:",omitempty"`
	Created  time.Time
	Expires  time.Time `json:",omitempty"`
	Actions  map[string]string
	path     string
}

func (n *Notification) ToStandardFormat() *respond.StandardFormat {
	return &respond.StandardFormat{
		Self:     n.path,
		Type:     "notification",
		Title:    n.Subject,
		Comment:  n.Body,
		IconName: n.IconName,
		OnPost:   n.Actions["default"],
		OnDelete: "Dismiss",
		Data:     n,
	}
}

func (n *Notification) haveDefaultAction() bool {
	_, ok := n.Actions["default"]
	return ok
}

func notificationSelf(id uint32) string {
	return fmt.Sprintf("/notification/%d", id)
}
