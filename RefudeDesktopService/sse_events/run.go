package sse_events

import (
	"fmt"
	"net"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/respond"
)

type Event struct {
	Type string
	Path string
}

var Publish = make(chan *Event)
var register = make(chan chan *Event)
var cancel = make(chan chan *Event)
var subscriptions = make(map[chan *Event]bool)

func Run() {
	for {
		select {
		case e := <-Publish:
			for c, _ := range subscriptions {
				// Concurrently, in case a recipient blocks.
				// We are aware that this may cause events to to become out-of-order
				cCopy := c
				go func() { cCopy <- e }()
			}
		case c := <-register:
			if subscriptions[c] {
				panic("Same channel in Subscribe twice")
			}
			subscriptions[c] = true
		case c := <-cancel:
			if subscriptions[c] {
				delete(subscriptions, c)
				close(c)
			}
		}
	}
}

const chunkTemplate = "%x\r\n" + // chunk length in hex
	"data:%s\n" +
	"\n" +
	"\r\n"

func write(conn net.Conn, msg string) bool {
	// TODO set some deadline...
	_, err := conn.Write([]byte(msg))
	return err == nil
}

/**
 *
 */
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Serving")
	if r.Method != "GET" {
		respond.NotAllowed(w)
	} else {
		fmt.Println("Serving events request")

		var watchedTypes = determineWatchedTypes(r.URL.Query()["type"])

		var ctx = r.Context()
		var subscription = make(chan *Event)
		register <- subscription

		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		w.(http.Flusher).Flush()
		fmt.Println("Header sent")
		fmt.Fprint(w, "\n")
		w.(http.Flusher).Flush()
		for {
			select {
			case <-ctx.Done():
				cancel <- subscription
			case ev := <-subscription:
				if ev == nil {
					fmt.Println("Returnerer")
					return
				} else {
					if watchedTypes[ev.Type] {
						fmt.Fprintf(w, "data:%s:%s\n\n", ev.Type, ev.Path)
						w.(http.Flusher).Flush()
					}
				}
			}
		}
	}
}

var allTypes = map[string]bool{"notification": true, "status_item": true, "power_device": true}

func determineWatchedTypes(types []string) map[string]bool {
	if len(types) == 0 {
		return allTypes
	} else {
		var set = make(map[string]bool, 10)
		for _, _type := range types {
			if _type == "events" {
				set["notification"] = true
			} else if allTypes[_type] {
				set[_type] = true
			} else {
				fmt.Println("Warn, unknown type:", _type) // error?
			}
		}
		return set
	}
}
