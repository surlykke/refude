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
	var event = eventSlot.Load().(*event)
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
	events <- &event{
		Sender:   sender,
		Title:    title,
		Message:  []string{message},
		IconName: iconName,
	}
}

func PublishGauge(sender, iconName string, gauge uint8) {
	events <- &event{
		Sender:   sender,
		IconName: iconName,
		Gauge:    &gauge,
	}
}

type event struct {
	Sender   string
	Gauge    *uint8   `json:",omitempty"`
	Title    string   `json:",omitempty"`
	Message  []string `json:",omitempty"`
	IconName string   `json:",omitempty"`
}

var events = make(chan *event)

var empty = &event{} // Used as a kind of nil
var eventSlot atomic.Value

func init() {
	eventSlot.Store(empty)
}

func Run() {
	var timeout = make(chan struct{})
	var buf = &buffer{}
	var currentExpires time.Time
	var timeoutPending = false

	for {
		var oldFirst = buf.first()

		select {
		case <-timeout:
			timeoutPending = false
			if time.Now().After(currentExpires) {
				buf.pop()
			}
		case ev := <-events:
			if buf.canMergeWithLast(ev) {
				buf.mergeWithLast(ev)
			} else if buf.len >= bufSize {
				fmt.Println("Buffer full, dropping event", ev)
			} else {
				buf.push(ev)
			}
		}

		if oldFirst != buf.first() {
			if buf.first() == nil {
				eventSlot.Store(empty)
			} else {
				eventSlot.Store(buf.first())
				currentExpires = time.Now().Add(6 * time.Second)
			}
			ss_events.Publish <- &ss_events.Event{Type: "osd", Path: "/osd"}
		}

		if buf.first() != nil && !timeoutPending {
			// schedule a timeout
			go func() {
				time.Sleep(currentExpires.Sub(time.Now()) + 10*time.Millisecond)
				timeout <- struct{}{}
			}()

			timeoutPending = true
		}
	}

}
