package watch

import (
	"fmt"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/pubsub"
	"github.com/surlykke/RefudeServices/lib/respond"
)

var events = pubsub.MakePublisher[string]()

func PublishStream(subscription *pubsub.Subscription[string]) {
	for {
		events.Publish(subscription.Next())
	}
}

func Publish(evt string) {
	events.Publish(evt)
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respond.NotAllowed(w)
	} else if r.URL.Path == "/watch" {
		var subscription = events.Subscribe()

		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.(http.Flusher).Flush()

		if _, err := fmt.Fprintf(w, "data:%s\n\n", ""); err != nil {
			return
		}
		w.(http.Flusher).Flush()

		for {
			if _, err := fmt.Fprintf(w, "data:%s\n\n", subscription.Next()); err != nil {
				return
			}
			w.(http.Flusher).Flush()
		}

	} else {
		respond.NotFound(w)
	}
}
