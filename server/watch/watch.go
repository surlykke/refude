package watch

import (
	"fmt"
	"net/http"

	"github.com/surlykke/refude/server/lib/pubsub"
)

type event struct {
	event string
	data  string
}

var events = pubsub.MakePublisher[event]()

func Publish(evt string, data string) {
	events.Publish(event{evt, data})
}

func ResourceChanged(path string) {
	Publish("resourceChanged", path)
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var subscription = events.Subscribe()

	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.(http.Flusher).Flush()

	for {
		var evt = subscription.Next()
		if _, err := fmt.Fprintf(w, "event:%s\n", evt.event); err != nil {
			return
		} else if _, err := fmt.Fprintf(w, "data:%s\n\n", evt.data); err != nil {
			return
		}
		w.(http.Flusher).Flush()
	}

}
