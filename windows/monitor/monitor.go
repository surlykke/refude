package monitor

import (
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/resource"
)

type MonitorData struct {
	X, Y     int
	W, H     int
	Wmm, Hmm int
	Name     string
	Primary  bool
}

func (this *MonitorData) Id() string {
	return this.Name
}

func (this *MonitorData) Presentation() (title string, comment string, icon link.Href, profile string) {
	return this.Name, "", "", "screen"
}

func (this *MonitorData) Links(self, searchTerm string) link.List {
	return link.List{}
}

func MakeMonitorWrapper(m *MonitorData) resource.Wrapper {
	return resource.MakeWrapper[string]("/window/screen/" + m.Name, m, "")
}



