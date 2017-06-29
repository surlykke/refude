package main

import (
	"net/http"
	"github.com/surlykke/RefudeServices/lib/resource"
)

type Notification struct {
	Id            uint32
	Sender        string
	Subject       string
	Body          string
	Actions       map[string]string
	RelevanceHint int
}

func NotificationPOST(this *resource.Resource, w http.ResponseWriter, r *http.Request) {
	action := resource.GetSingleQueryParameter(r, "action", "default")
	conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE + ".ActionInvoked", this.Data.(Notification).Id, action)
	w.WriteHeader(http.StatusAccepted)
}

func NotificationDELETE(this *resource.Resource,  w http.ResponseWriter, r *http.Request) {
	close(r.URL.Path, "", Dismissed)
}
