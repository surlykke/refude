package buffer

import (
	"fmt"
	"time"
)

type Event struct {
	Sender   string
	Gauge    *uint8   `json:",omitempty"`
	Title    string   `json:",omitempty"`
	Message  []string `json:",omitempty"`
	IconName string   `json:",omitempty"`
}

const bufSize = 64

var (
	data           [bufSize]*Event
	start          uint8
	size           uint8
	CurrentTimeout time.Time
	CurrentUpdated bool
)

func Current() *Event {
	if size == 0 {
		return nil
	} else {
		return data[start]
	}
}

func last() uint8 {
	return (start + size - 1) % bufSize
}

func isAGaugeEvent(e *Event) bool {
	return e.Gauge != nil
}

func Push(e *Event) {
	if isAGaugeEvent(e) {
		// We drop gauge events if there is currently something showing which is not a gauge event or a gauge event for something else
		if size == 0 || size == 1 && isAGaugeEvent(Current()) && Current().Sender == e.Sender {
			data[start] = e
			size = 1 // add or overwrite first
			CurrentTimeout = time.Now().Add(2 * time.Second)
			CurrentUpdated = true
		}
	} else {
		fmt.Println("e:", e, ", last():", last(), ", data[last()]", data[last()])
		if size > 0 && !isAGaugeEvent(data[last()]) && data[last()].Sender == e.Sender && data[last()].Title == e.Title && len(data[last()].Message) < 3 {
			data[last()] = &Event{
				Sender:   e.Sender,
				Title:    e.Title,
				Message:  append(e.Message, data[last()].Message...),
				IconName: e.IconName,
			}
		} else if size >= bufSize {
			fmt.Println("Buffer full, dropping osd event")
		} else {
			size++
			data[last()] = e
		}

		if size == 1 {
			CurrentTimeout = time.Now().Add(6 * time.Second)
			CurrentUpdated = true
		}
	}

	fmt.Println("After push, start, size:", start, size)
}

func Pop() {
	if size == 0 {
		fmt.Println("Pop from empty buffer")
	} else {
		data[start] = nil // For the benefit of gc
		start = (start + 1) % bufSize
		size--

		CurrentUpdated = true
		if size > 0 { // We do not queue gauge events, so new first must be a message
			CurrentTimeout = time.Now().Add(6 * time.Second)
		}
	}

	fmt.Println("After pop, start, size:", start, size)
}
