// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package notifications

import (
	"strconv"
	"time"

	"github.com/surlykke/refude/internal/lib/resource"
)

type Urgency uint8

const (
	Low Urgency = iota
	Normal
	Critical
)

var (
	LowBytes      = []byte(`"low"`)
	NormalBytes   = []byte(`"normal"`)
	CriticalBytes = []byte(`"critical"`)
)

func (u Urgency) MarshalJSON() ([]byte, error) {
	switch u {
	case Low:
		return LowBytes, nil
	case Normal:
		return NormalBytes, nil
	case Critical:
		return CriticalBytes, nil
	default:
		panic("unknown urgency")
	}
}

type UnixTime time.Time // Behaves like Time, but json-marshalls to milliseconds since epoch

func (ut UnixTime) MarshalJSON() ([]byte, error) {
	var buf = make([]byte, 0, 22)
	buf = strconv.AppendInt(buf, time.Time(ut).UnixMilli(), 10)
	return buf, nil
}

type Notification struct {
	resource.Base
	NotificationId uint32
	Body           string
	Sender         string
	Created        time.Time
	Expires        time.Time
	Expired        bool // Set once expiration has been published on dbus
	Urgency        Urgency
	NActions       map[string]string `json:"actions"`
	Hints          map[string]any
	IconName       string
}

func (n *Notification) OmitFromSearch() bool {
	return n.Expired
}

func (n *Notification) DoPost(action string) (bool, error) {
	if _, ok := n.NActions[action]; ok {
		if err := conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".ActionInvoked", n.NotificationId, action); err != nil {
			return false, err
		} else {
			return true, nil
		}
	} else {
		return false, nil
	}
}

func (n *Notification) DoDelete() error {
	dismissedNotifications <- n.NotificationId 
	return nil
}
