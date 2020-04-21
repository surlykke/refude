package osd

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/ss_events"

	"github.com/surlykke/RefudeServices/lib/respond"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var event = eventStore.Load().(*event)
	if r.URL.Path != "/osd" || event == empty {
		respond.NotFound(w)
	} else {
		respond.AsJson(w, r, event)
	}
}

func PublishMessage(sender, title, message, iconName string) {
	if iconName == "" {
		iconName = "dialog-information"
	}
	var event = &event{
		Sender:   sender,
		Title:    title,
		Message:  []string{message},
		IconName: iconName,
	}
	select {
	case events <- event: // all is well
	default:
		fmt.Println("event buffer full. Dropping", event)
	}
}

type event struct {
	Sender   string
	Title    string
	Message  []string `json:",omitempty"`
	IconName string
}

var events = make(chan *event, 50)

const showTime = 6 * time.Second

var empty = &event{} // Used as a kind of nil
var eventStore atomic.Value

func currentEvent() *event {
	return eventStore.Load().(*event)
}

func setCurrentEvent(event *event) {
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

func canAmend(current, next *event) bool {
	return current.Sender == next.Sender &&
		current.Title == next.Title &&
		len(current.Message) > 0 &&
		len(current.Message) < 3 &&
		len(next.Message) > 0 // Will never be > 1

}

func amend(current, next *event) *event {
	return &event{
		Sender:   current.Sender,
		Title:    current.Title,
		Message:  append(next.Message, current.Message...),
		IconName: next.IconName,
	}
}

func Run() {
	var next *event = nil
	var currentExpires time.Time
	for {
		var current = currentEvent()
		if current == empty { // Then next will be nil
			current = <-events
			currentExpires = time.Now().Add(showTime)
			setCurrentEvent(current)
			go scheduleTimeout(currentExpires)
		} else {
			if next == nil {
				select {
				case <-timeout:
					if currentExpires.Before(time.Now()) {
						setCurrentEvent(empty)
					} else {
						go scheduleTimeout(currentExpires)
					}
				case next = <-events:
					if canAmend(current, next) {
						setCurrentEvent(amend(current, next))
						currentExpires = time.Now().Add(showTime)
						next = nil
					}
				}
			} else {
				<-timeout
				if currentExpires.Before(time.Now()) {
					setCurrentEvent(empty)
				} else {
					go scheduleTimeout(currentExpires)
				}
			}
		}

	}
}
