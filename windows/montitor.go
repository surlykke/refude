package windows

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/respond"
)

type Monitor struct {
	respond.Links `json:"_links"`
	MonitorData
}

func monitorDataList2Monitors(monitorDataList []MonitorData) []*Monitor {
	var monitors = make([]*Monitor, len(monitorDataList), len(monitorDataList))
	for i, md := range monitorDataList {
		monitors[i] = &Monitor{
			respond.Links{}, // FIXME
			md,
		}
	}

	return monitors
}

func (m *Monitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, m)
	} else {
		respond.NotAllowed(w)
	}
}
