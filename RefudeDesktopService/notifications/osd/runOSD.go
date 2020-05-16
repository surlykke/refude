package osd

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/watch"
)

func PublishMessage(id uint32, sender, title, message, iconName string) {
	if iconName == "" {
		iconName = "dialog-information"
	}
	events <- &Event{
		notificationId: id,
		Sender:         sender,
		Title:          title,
		Message:        []string{message},
		IconName:       iconName,
	}
}

func PublishGauge(id uint32, sender, iconName string, gauge uint8) {
	events <- &Event{
		notificationId: id,
		Sender:         sender,
		IconName:       iconName,
		Gauge:          &gauge,
	}
}

var events = make(chan *Event)

var empty = &Event{} // Used as a kind of nil
var eventSlot atomic.Value

func CurrentlyShowing() *Event {
	var e = eventSlot.Load().(*Event)
	if e == empty {
		return nil
	} else {
		return e
	}
}

func init() {
	eventSlot.Store(empty)
}

func RunOSD() {
	var timeout = make(chan struct{})
	var timeoutPending = false

	for {
		select {
		case <-timeout:
			timeoutPending = false
			if time.Now().After(currentTimeout) {
				pop()
			}
		case ev := <-events:
			push(ev)
		}

		if first() != nil && !timeoutPending {
			go func() {
				time.Sleep(currentTimeout.Sub(time.Now()) + 10*time.Millisecond)
				timeout <- struct{}{}
			}()
			timeoutPending = true
		}

	}

}

type Event struct {
	notificationId uint32
	Sender         string
	Gauge          *uint8   `json:",omitempty"`
	Title          string   `json:",omitempty"`
	Message        []string `json:",omitempty"`
	IconName       string   `json:",omitempty"`
}

const bufSize = 64

var (
	data           [bufSize]*Event
	size           uint8
	currentTimeout time.Time
	currentUpdated bool
)

func first() *Event {
	if size == 0 {
		return nil
	} else {
		return data[0]
	}
}

func last() *Event {
	if size == 0 {
		return nil
	} else {
		return data[size-1]
	}
}

func isAGaugeEvent(e *Event) bool {
	return e.Gauge != nil
}

func push(e *Event) {
	if !replaceEvent(e) {
		if isAGaugeEvent(e) {
			// We drop gauge events if there is currently something showing which is not a gauge event or a gauge event for something else
			if size == 0 || size == 1 && isAGaugeEvent(first()) && first().Sender == e.Sender {
				data[0] = e
				size = 1 // add or overwrite first
				updateCurrent()
			}
		} else {
			if size > 0 && !isAGaugeEvent(last()) && last().Sender == e.Sender && last().Title == e.Title && len(last().Message) < 3 {
				data[size-1] = &Event{
					Sender:   e.Sender,
					Title:    e.Title,
					Message:  append(e.Message, data[size-1].Message...),
					IconName: e.IconName,
				}
			} else if size >= bufSize {
				fmt.Println("Buffer full, dropping osd event")
			} else {
				data[size] = e
				size++
			}

			if size == 1 {
				updateCurrent()
			}
		}
	}
}

func replaceEvent(e *Event) bool {
	for i := uint8(0); i < size; i++ {
		if e.notificationId == data[i].notificationId {
			data[i] = e
			if i == 0 {
				updateCurrent()
			}
			return true
		}
	}
	return false
}

func updateCurrent() {
	if first() == nil {
		eventSlot.Store(empty)
	} else {
		eventSlot.Store(first())
		if isAGaugeEvent(first()) {
			currentTimeout = time.Now().Add(2 * time.Second)
		} else {
			currentTimeout = time.Now().Add(6 * time.Second)
		}
	}
	watch.SomethingChanged("/notification/osd")
}

func pop() {
	if size == 0 {
		fmt.Println("Pop from empty buffer")
	} else {
		for i := uint8(1); i < size; i++ {
			data[i-1] = data[i]
		}
		size--
		data[size] = nil // For the benefit of gc

		updateCurrent()
	}
}
