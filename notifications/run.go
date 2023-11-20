// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package notifications

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/watch"
)

var notificationExpireryHints = make(chan struct{})

func Run() {
	go DoDBus()
}

func removeNotification(id uint32, reason uint32) {
	if n, ok := Notifications.Get(strconv.Itoa(int(id))); ok && !n.Deleted {
		var copy = *n
		copy.Deleted = true
		Notifications.Update(&copy)
		conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, reason)
		calculateFlash()
	}
}

type Flash map[string]string

func ServeFlash(w http.ResponseWriter, r *http.Request) {
	flashLock.Lock()
	var f = flash
	flashLock.Unlock()
	if f == nil {
		respond.NotFound(w)
	} else {
		var f = make(Flash)

		f["subject"] = flash.Title
		f["body"] = flash.Comment
		f["iconFilePath"] = icons.FindIconPath(flash.iconName, 64)
		switch flash.Urgency {
		case Critical:
			f["urgency"] = "critical"
		case Normal:
			f["urgency"] = "normal"
		case Low:
			f["urgency"] = "low"
		}
		respond.AsJson(w, f)
	}
}

var flash *Notification
var flashLock sync.Mutex

func calculateFlash() {
	var calculatedFlash *Notification = nil
	var now = time.Now()
	for _, n := range Notifications.GetAll() {
		if n.Deleted {
			continue
		}

		if n.Urgency == Critical {
			calculatedFlash = n
			break
		} else if n.Urgency == Normal {
			if calculatedFlash == nil || calculatedFlash.Urgency < Normal {
				if now.Before(time.Time(n.Created).Add(6 * time.Second)) {
					calculatedFlash = n
				}
			}
		} else { /* n.Urgency == Low */
			if calculatedFlash == nil && now.Before(time.Time(n.Created).Add(2*time.Second)) {
				calculatedFlash = n
			}
		}
	}

	if calculatedFlash != nil {
		if calculatedFlash.Urgency != Critical {
			var timeout = time.Time(calculatedFlash.Created).Sub(time.Now()) + 50*time.Millisecond
			if calculatedFlash.Urgency == Normal {
				timeout = timeout + 6*time.Second
			} else {
				timeout = timeout + 2*time.Second
			}
			time.AfterFunc(timeout, calculateFlash)
		}
	}

	if calculatedFlash != flash {
		flashLock.Lock()
		flash = calculatedFlash
		flashLock.Unlock()
		watch.ResourceChanged("/flash")
	}

}
