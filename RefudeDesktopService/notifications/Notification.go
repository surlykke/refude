// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"io"
	"net/http"
	"time"

	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/serialize"
)

const NotificationMediaType resource.MediaType = "application/vnd.org.refude.desktopnotification+json"

type Notification struct {
	resource.GenericResource
	Id      uint32
	Sender  string
	Subject string
	Body    string
	Created time.Time
	Expires time.Time `json:",omitempty"`
}

func (n *Notification) removeAfter(duration time.Duration) {
	time.AfterFunc(duration, func() { removals <- removal{n.Id, Expired} })
}

func (nc *Notification) DELETE(w http.ResponseWriter, r *http.Request) {
	removals <- removal{id: nc.Id, reason: Dismissed}
}

func notificationSelf(id uint32) resource.StandardizedPath {
	return resource.Standardizef("/notifications/%d", id)
}

func (n *Notification) WriteBytes(w io.Writer) {
	n.GenericResource.WriteBytes(w)
	serialize.UInt32(w, n.Id)
	serialize.String(w, n.Sender)
	serialize.String(w, n.Subject)
	serialize.String(w, n.Body)
	serialize.UInt64(w, uint64(n.Created.UnixNano()))
	serialize.UInt64(w, uint64(n.Expires.UnixNano()))
}
