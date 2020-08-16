package notifications

import (
	"net/http"
	"time"

	"github.com/surlykke/RefudeServices/lib/respond"
)

type eventType string

const (
	critical eventType = "critical"
	normal             = "normal"
	gauge              = "gauge"
	none               = "none"
)

type osdEvent struct {
	respond.Links `json:"_links"`
	Type          eventType
	Expires       time.Time
	Sender        string
	Gauge         uint8    `json:",omitempty"`
	Subject       string   `json:",omitempty"`
	Body          []string `json:",omitempty"`
	iconName      string
}

func (e osdEvent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, e)
	} else {
		respond.NotAllowed(w)
	}
}
