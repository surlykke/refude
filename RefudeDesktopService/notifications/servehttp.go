package notifications

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

var notificationPathPattern = regexp.MustCompile("^/notification/(\\d+)$")

func Handler(r *http.Request) http.Handler {
	if r.URL.Path == "/notification/osd" {
		var e = box2.Load().(osdEvent)
		if e.Type == none {
			return nil
		} else {
			return e
		}
	} else if r.URL.Path == "/notifications" {
		return Collect()
	} else if matches := notificationPathPattern.FindStringSubmatch(r.URL.Path); matches != nil {
		fmt.Print("Submatch: ", matches)
		if id, err := strconv.Atoi(matches[1]); err == nil {
			fmt.Println("id:", id)
			if notification := getNotification(uint32(id)); notification != nil {
				return notification
			}
		}
	}

	return nil
}

func Collect() respond.Links {
	var notifications = box.Load().([]*Notification)
	var links = make(respond.Links, 0, len(notifications))
	for _, notification := range notifications {
		links = append(links, notification.Link())
	}
	return links
}

func DesktopSearch(term string, baserank int) respond.Links {
	var notifications = box.Load().([]*Notification)
	var links = make(respond.Links, 0, len(notifications))
	for _, notification := range notifications {
		if _, ok := notification.Actions["default"]; ok {
			var link = notification.Link()
			if link.Rank, ok = searchutils.Rank(notification.Subject, term, baserank); !ok {
				link.Rank, ok = searchutils.Rank(notification.Body, term, baserank+100)
			}
			if ok {
				links = append(links, link)
			}
		}
	}
	return links
}

func AllPaths() []string {
	var notifications = box.Load().([]*Notification)
	var paths = make([]string, 0, len(notifications)+2)
	for _, n := range notifications {
		paths = append(paths, n.self)
	}
	paths = append(paths, "/notifications")
	paths = append(paths, "/notification/osd")
	return paths
}
