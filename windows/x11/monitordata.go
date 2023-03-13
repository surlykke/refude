package x11

import "github.com/surlykke/RefudeServices/lib/resource"

type MonitorData struct {
	resource.BaseResource
	X, Y     int
	W, H     int
	Wmm, Hmm int
	Primary  bool
}


