// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package power

import (
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/server"
	"net/http"
	"strings"
	"sync"
)

const DeviceMediaType resource.MediaType = "application/vnd.org.refude.upowerdevice+json"

type Device struct {
	resource.AbstractResource
	Id               string
	NativePath       string
	Vendor           string
	Model            string
	Serial           string
	UpdateTime       uint64
	Type             string
	PowerSupply      bool
	HasHistory       bool
	HasStatistics    bool
	Online           bool
	Energy           float64
	EnergyEmpty      float64
	EnergyFull       float64
	EnergyFullDesign float64
	EnergyRate       float64
	Voltage          float64
	TimeToEmpty      int64
	TimeToFull       int64
	Percentage       float64
	IsPresent        bool
	State            string
	IsRechargeable   bool
	Capacity         float64
	Technology       string
	DisplayDevice    bool
}

func deviceType(index uint32) string {
	var devType = []string{"Unknown", "Line Power", "Battery", "Ups", "Monitor", "Mouse", "Keyboard", "Pda", "Phone"}
	if index < 0 || index > 8 {
		index = 0
	}
	return devType[index]
}

func deviceState(index uint32) string {
	var devState = []string{"Unknown", "Charging", "Discharging", "Empty", "Fully charged", "Pending charge", "Pending discharge"}
	if index < 0 || index > 6 {
		index = 0
	}
	return devState[index]
}

func deviceTecnology(index uint32) string {
	var devTecnology = []string{"Unknown", "Lithium ion", "Lithium polymer", "Lithium iron phosphate", "Lead acid", "Nickel cadmium", "Nickel metal hydride"}
	if index < 0 || index > 6 {
		index = 0
	}
	return devTecnology[index]
}

type PowerCollection struct {
	mutex   sync.Mutex
	devices map[string]*Device
	session *Session
	server.CachingJsonGetter
	server.PatchNotAllowed
	server.DeleteNotAllowed
}

func (*PowerCollection) HandledPrefixes() []string {
	return []string{"/device", "/session"}
}

func MakePowerCollection() *PowerCollection {
	var dc = &PowerCollection{}
	dc.CachingJsonGetter = server.MakeCachingJsonGetter(dc)
	dc.devices = make(map[string]*Device)
	return dc
}

func (pc *PowerCollection) GetSingle(r *http.Request) interface{} {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	var path = r.URL.Path
	if strings.HasPrefix(path, "/device/") {
		if device, ok := pc.devices[path[len("/device/"):]]; ok {
			return device
		}
	} else if path == "/session" {
		return pc.session
	}
	return nil
}

func (pc *PowerCollection) GetCollection(r *http.Request) []interface{} {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	if r.URL.Path == "/devices" {
		var result = make([]interface{}, 0, len(pc.devices))
		for _, device := range pc.devices {
			result = append(result, device)
		}
		return result
	} else {
		return nil
	}
}

func (pc *PowerCollection) POST(w http.ResponseWriter, r *http.Request) {
	if res := pc.GetSingle(r); res == nil {
		w.WriteHeader(http.StatusNotFound)
	} else if session, ok := res.(*Session); !ok {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		var actionId = requests.GetSingleQueryParameter(r, "action", "Suspend")
		if action, ok := session.ResourceActions[actionId]; ok {
			action.Executer()
			w.WriteHeader(http.StatusAccepted)
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
	}
}
