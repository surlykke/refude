// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package watch

import (
	"fmt"
	"net/http"

	"github.com/surlykke/refude/internal/applications"
	"github.com/surlykke/refude/internal/browser"
	"github.com/surlykke/refude/internal/file"
	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/internal/notifications"
	"github.com/surlykke/refude/internal/power"
	"github.com/surlykke/refude/internal/wayland"
	"github.com/surlykke/refude/pkg/pubsub"
)

type event struct {
	event string
	data  string
}

var aggregatedEvents = pubsub.MakePublisher[entity.Event]()

func follow(events *pubsub.Publisher[entity.Event]) {
	var subscription = events.Subscribe()
	for {
		aggregatedEvents.Publish(subscription.Next())
	}
}

func Run() {
	go follow(applications.AppMap.Events)
	go follow(applications.MimeMap.Events)
	go follow(wayland.WindowMap.Events)
	go follow(notifications.NotificationMap.Events)
	go follow(browser.BookmarkMap.Events)
	go follow(browser.TabMap.Events)
	go follow(power.DeviceMap.Events)
	go follow(file.FileMap.Events)
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var subscription = aggregatedEvents.Subscribe()

	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.(http.Flusher).Flush()

	for {
		var evt = subscription.Next()
		if _, err := fmt.Fprintf(w, "event:%s\n", evt.Event); err != nil {
			return
		} else if _, err := fmt.Fprintf(w, "data:%s\n\n", evt.Data); err != nil {
			return
		}
		w.(http.Flusher).Flush()
	}

}
