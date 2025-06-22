// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package notifications

import (
	"strconv"
	"time"

	"github.com/surlykke/RefudeServices/lib/entity"
	"github.com/surlykke/RefudeServices/lib/icon"
	"github.com/surlykke/RefudeServices/lib/response"
)

type Urgency uint8

const (
	Low      Urgency = 0
	Normal           = 1
	Critical         = 2
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
	entity.Base
	NotificationId uint32
	Body           string
	Sender         string
	Created        time.Time
	Expires        time.Time
	Deleted        bool
	Urgency        Urgency
	NActions       map[string]string `json:"actions"`
	Hints          map[string]interface{}
	iconName       icon.Name
	IconSize       uint32 `json:",omitempty"`
}

func (n *Notification) Expired() bool {
	return time.Now().After(time.Time(n.Expires))
}

func (n *Notification) SoftExpired() bool {
	return n.Urgency == Normal && n.Created.Add(10*time.Second).Before(time.Now()) ||
		n.Urgency == Low && n.Created.Add(2*time.Second).Before(time.Now())
}

func (this *Notification) OmitFromSearch() bool {
	return this.Deleted || this.Expired() || (this.NActions["default"] == "" && this.SoftExpired())
}

func (n *Notification) DoPost(action string) response.Response {
	if action == "" && len(n.Actions) > 0 {
		action = n.Actions[0].Id
	}
	if _, ok := n.NActions[action]; ok {
		if err := conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".ActionInvoked", n.NotificationId, action); err != nil {
			return response.ServerError(err)
		} else {
			return response.Accepted()
		}
	} else {
		return response.NotFound()
	}
}

func (n *Notification) DoDelete() response.Response {
	removeNotification(n.NotificationId, Dismissed)
	return response.Ok()
}
