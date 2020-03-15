package backlight

import (
	"net/http"
	"time"

	"github.com/surlykke/RefudeServices/lib/respond"
)

type Device struct {
	Id            string
	BrightnessPct uint8
	Updated       time.Time
	maxBrightness uint64
	brightness    uint64
}

func (d *Device) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, d.ToStandardFormat())
	} else {
		respond.NotAllowed(w)
	}
}

func (dev *Device) ToStandardFormat() *respond.StandardFormat {
	return &respond.StandardFormat{
		Self:  "/backlight/" + dev.Id,
		Type:  "backlight_device",
		Title: dev.Id,
		Data:  dev,
	}
}

type DeviceMap map[string]*Device
