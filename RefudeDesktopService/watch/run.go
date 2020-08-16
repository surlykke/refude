package watch

import (
	"fmt"
	"net"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/respond"
)

type subscription chan string

var changedPaths = make(chan string)
var subscriptions = make(chan subscription)
var cancellations = make(chan subscription)
var subscriptionSet = make(map[subscription]bool)

func SomethingChanged(path string) {
	changedPaths <- path
}

func subscribe() subscription {
	var s = make(subscription)
	subscriptions <- s
	return s
}

func cancel(s subscription) {
	cancellations <- s
}

func Run() {
	for {
		select {
		case path := <-changedPaths:
			for s := range subscriptionSet {
				// Concurrently, in case a recipient blocks.
				sCopy := s
				go func() { sCopy <- path }()
			}
		case s := <-subscriptions:
			subscriptionSet[s] = true
		case s := <-cancellations:
			delete(subscriptionSet, s)
			close(s)
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

func Handler(r *http.Request) http.Handler {
	if r.URL.Path == "/watch" {
		return http.HandlerFunc(ServeHTTP)
	} else {
		return nil
	}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respond.NotAllowed(w)
	} else if r.URL.Path == "/watch" {
		var done = r.Context().Done()
		var s = subscribe()

		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.(http.Flusher).Flush()

		fmt.Fprintf(w, "data:%s\n\n", "")
		w.(http.Flusher).Flush()
		for {
			select {
			case <-done:
				cancellations <- s
				return
			case path := <-s:
				fmt.Fprintf(w, "data:%s\n\n", path)
				w.(http.Flusher).Flush()
			}
		}
	} else {
		respond.NotFound(w)
	}
}
