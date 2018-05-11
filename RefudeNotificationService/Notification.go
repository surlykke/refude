// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"time"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/resource"
)

const NotificationMediaType mediatype.MediaType = "application/vnd.org.refude.desktopnotification+json"

type Notification struct {
	resource.Self
	Id            uint32
	Sender        string
	Subject       string
	Body          string
	Actions       map[string]string
	RelevanceHint int
	Expires       *time.Time `json:",omitempty"`
}

