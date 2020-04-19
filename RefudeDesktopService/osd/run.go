package osd

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/ss_events"

	"github.com/surlykke/RefudeServices/lib/respond"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var event = eventStore.Load().(*Event)
	if r.URL.Path != "/osd" || event == empty {
		respond.NotFound(w)
	} else {
		respond.AsJson(w, r, event)
	}
}

type Event struct {
	Sender   string
	Title    string
	Body     string
	IconName string
	ends     time.Time
}

var Events = make(chan *Event)

const showTime = 6 * time.Second

var empty = &Event{} // Used as a kind of nil
var eventStore atomic.Value

func currentEvent() *Event {
	return eventStore.Load().(*Event)
}

func setCurrentEvent(event *Event) {
	eventStore.Store(event)
	ss_events.Publish <- &ss_events.Event{Type: "osd", Path: "/osd"}
}

func init() {
	eventStore.Store(empty)
}

var timeout = make(chan struct{})

// Call concurrently
func scheduleTimeout(t time.Time) {
	time.Sleep(t.Sub(time.Now()) + 10*time.Millisecond)
	timeout <- struct{}{}
}

func canAmend(current, next *Event) bool {
	return false // FIXME
}

func amend(current, next *Event) *Event {
	return &(*current) // Fixme
}

func Run() {
	var next *Event = nil
	for {
		var current = currentEvent()
		if current == empty {
			current = <-Events
			current.ends = time.Now().Add(showTime)
			setCurrentEvent(current)
			go scheduleTimeout(current.ends)
		} else {
			if next == nil {
				select {
				case <-timeout:
					if current.ends.Before(time.Now()) {
						setCurrentEvent(empty)
					} else {
						go scheduleTimeout(current.ends)
					}
				case next = <-Events:
					if canAmend(current, next) {
						setCurrentEvent(amend(current, next))
						next = nil
					}
				}
			} else {
				<-timeout
				if current.ends.Before(time.Now()) {
					setCurrentEvent(empty)
				} else {
					go scheduleTimeout(current.ends)
				}
			}
		}

	}
}
