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

	"github.com/surlykke/RefudeServices/lib/resource"
)

type Notification struct {
	resource.GeneralTraits
	Id      uint32
	Sender  string
	Subject string
	Body    string
	Created time.Time
	Expires time.Time `json:",omitempty"`
}

func (nc *Notification) DELETE(w http.ResponseWriter, r *http.Request) {
	removals <- removal{id: nc.Id, reason: Dismissed}
}

func notificationSelf(id uint32) string {
	return fmt.Sprintf("/notification/%d", id)
}
