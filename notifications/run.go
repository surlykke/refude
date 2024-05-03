// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package notifications

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/resourcerepo"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/watch"
)

func Run() {
	go DoDBus()
}

func removeNotification(id uint32, reason uint32) {
	if n, ok := resourcerepo.GetTyped[*Notification](fmt.Sprintf("/notification/%d", id)); ok && !n.Deleted {
		var copy = *n
		copy.Deleted = true
		resourcerepo.Update(&copy)
		conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, reason)
		calculateFlash()
	}
}

func Search(list *resource.RRList, term string) {
	for _, n := range resourcerepo.GetTypedByPrefix[*Notification]("/notification/") {
		if len(term) > 2 {
			if _, ok := n.NActions["default"]; ok {
				if rnk := searchutils.Match(term, n.Title, "notification"); rnk >= 0 {
					*list = append(*list, resource.RankedResource{Res: n, Rank: rnk})
				}
			}
		}
	}
}

type Flash struct {
	valid        bool
	Subject      string `json:"subject"`
	Body         string `json:"body"`
	IconFilePath string `json:"iconFilePath"`
	Urgency      Urgency
}

func MakeFlash(n *Notification) Flash {
	return Flash{valid: true, Subject: n.Title, Body: n.Comment, IconFilePath: icons.FindIconPath(n.iconName, 64), Urgency: n.Urgency}
}

var currentFlash atomic.Pointer[Flash]

func init() {
	currentFlash.Store(&Flash{})
}

func ServeFlash(w http.ResponseWriter, r *http.Request) {
	var f = currentFlash.Load()
	if f.valid {
		respond.AsJson(w, f)
	} else {
		respond.NotFound(w)
	}
}

func calculateFlash() {
	var now = time.Now()
	notifications := resourcerepo.GetTypedByPrefix[*Notification]("/notification/")
	var pos = -1
	for i := len(notifications) - 1; i >= 0; i-- {
		var n = notifications[i]
		if n.Deleted {
			continue
		}

		if n.Urgency == Critical {
			pos = i
			break
		} else if n.Urgency == Normal {
			if pos == -1 || notifications[i].Urgency < Normal {
				if now.Before(time.Time(n.Created).Add(6 * time.Second)) {
					pos = i
				}
			}
		} else { /* n.Urgency == Low */
			if pos == -i && now.Before(time.Time(n.Created).Add(2*time.Second)) {
				pos = i
			}
		}
	}

	var newFlash Flash
	if pos > -1 {
		var n = notifications[pos]
		newFlash = MakeFlash(n)
		if newFlash.Urgency != Critical {
			var timeout = time.Time(n.Created).Sub(time.Now()) + 50*time.Millisecond
			if newFlash.Urgency == Normal {
				timeout = timeout + 6*time.Second
			} else {
				timeout = timeout + 2*time.Second
			}
			time.AfterFunc(timeout, calculateFlash)
		}
	}

	var oldFlash = currentFlash.Swap(&newFlash)

	if *oldFlash != newFlash {
		watch.ResourceChanged("/flash")
	}

}
