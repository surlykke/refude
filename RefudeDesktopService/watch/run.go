package watch

import (
	"fmt"
	"net"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
)

type subscription struct {
	path string
	ch   chan struct{}
}

var changedPaths = make(chan string)
var subscriptions = make(chan subscription)
var cancellations = make(chan subscription)
var subscriptionMap = make(map[string][]chan struct{})

func SomethingChanged(path string) {
	changedPaths <- path
}

func subscribe(path string) subscription {
	var res = subscription{path, make(chan struct{})}
	subscriptions <- res
	return res
}

func cancel(s subscription) {
	cancellations <- s
}

func Run() {
	for {
		select {
		case path := <-changedPaths:
			for _, c := range subscriptionMap[path] {
				// Concurrently, in case a recipient blocks.
				cCopy := c
				go func() { cCopy <- struct{}{} }()
			}
		case s := <-subscriptions:
			subscriptionMap[s.path] = append(subscriptionMap[s.path], s.ch)
		case s := <-cancellations:
			var filtered = make([]chan struct{}, 0, len(subscriptionMap[s.path]))
			for _, ch := range subscriptionMap[s.path] {
				if ch != s.ch {
					filtered = append(filtered, ch)
				}
			}
			close(s.ch)
			subscriptionMap[s.path] = filtered
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

/**
 *
 */
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respond.NotAllowed(w)
	} else if r.URL.Path != "/watch" {
		respond.NotFound(w)
	} else if path := requests.GetSingleQueryParameter(r, "path", ""); path == "" {
		respond.UnprocessableEntity(w, fmt.Errorf("query parameter 'path' must be given"))
	} else {
		var done = r.Context().Done()
		var subscription = subscribe(path)

		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.(http.Flusher).Flush()

		fmt.Fprintf(w, "data:%s\n\n", path)
		w.(http.Flusher).Flush()
		for {
			select {
			case <-done:
				cancel(subscription)
				done = nil
			case _, ok := <-subscription.ch:
				if !ok {
					return
				} else {
					fmt.Fprintf(w, "data:%s\n\n", path)
					w.(http.Flusher).Flush()
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
