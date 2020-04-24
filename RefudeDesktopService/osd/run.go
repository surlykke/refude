package osd

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/osd/buffer"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/ss_events"
	"github.com/surlykke/RefudeServices/lib/respond"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var event = eventSlot.Load().(*buffer.Event)
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
	events <- &buffer.Event{
		Sender:   sender,
		Title:    title,
		Message:  []string{message},
		IconName: iconName,
	}
}

func PublishGauge(sender, iconName string, gauge uint8) {
	events <- &buffer.Event{
		Sender:   sender,
		IconName: iconName,
		Gauge:    &gauge,
	}
}

var events = make(chan *buffer.Event)

var empty = &buffer.Event{} // Used as a kind of nil
var eventSlot atomic.Value

func init() {
	eventSlot.Store(empty)
}

func Run() {
	var timeout = make(chan struct{})
	var timeoutPending = false

	for {
		fmt.Println("loop")
		buffer.CurrentUpdated = false

		select {
		case <-timeout:
			fmt.Println("Timeout")
			timeoutPending = false
			if time.Now().After(buffer.CurrentTimeout) {
				buffer.Pop()
			}
		case ev := <-events:
			fmt.Println("incoming")
			buffer.Push(ev)
		}

		fmt.Println("currentUpdated:", buffer.CurrentUpdated, "current():", buffer.Current())

		if buffer.CurrentUpdated {
			if buffer.Current() == nil {
				eventSlot.Store(empty)
			} else {
				eventSlot.Store(buffer.Current())
			}
			ss_events.Publish <- &ss_events.Event{Type: "osd", Path: "/osd"}
		}

		if buffer.Current() != nil && !timeoutPending {
			fmt.Println("Schedule timeout..")
			go func() {
				time.Sleep(buffer.CurrentTimeout.Sub(time.Now()) + 10*time.Millisecond)
				timeout <- struct{}{}
			}()
			timeoutPending = true
		}

	}

}
