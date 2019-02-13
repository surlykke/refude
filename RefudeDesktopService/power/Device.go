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
	var devTecnology = []string{"Unknown", "Lithium ion", "Lithium polymer", "Lithium iron phosphate", "Lead acid", "Nickel cadmium", "Nickel metal hydride" }
	if index < 0 || index > 6 {
		index = 0
	}
	return devTecnology[index]
}

type DevicesCollection struct {
	sync.Mutex
	server.JsonResponseCache
	devices map[string]*Device
}


func MakeDevicesCollection() *DevicesCollection {
	var dc = &DevicesCollection{}
	dc.JsonResponseCache = server.MakeJsonResponseCache(dc)
	dc.devices = make(map[string]*Device)
	return dc
}

func (dac DevicesCollection) GetResource(r *http.Request) (interface{}, error) {
	var path = r.URL.Path
	if path == "/devices" {
		var devices = make([]*Device, 0, len(dac.devices))

		var matcher, err = requests.GetMatcher(r);
		if err != nil {
			return nil, err
		}

		for _, device := range dac.devices {
			if matcher(device) {
				devices = append(devices, device)
			}
		}

		return devices, nil
	} else if strings.HasPrefix(path, "/device/") {
		if device, ok := dac.devices[path[len("/device/"):]]; ok {
			return device, nil
		} else {
			return nil, nil
		}
	} else {
		return nil, nil
	}

}