package session

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/respond"

	"github.com/godbus/dbus/v5"
)

func Handler(r *http.Request) http.Handler {
	if r.URL.Path == "/session/actions" {
		return Collect()
	} else if _, ok := actions[r.URL.Path]; ok {
		return Action(r.URL.Path)
	} else {
		return nil
	}
}

type Action string

func (a Action) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var sf = actions[string(a)]
	if r.Method == "GET" {
		respond.AsJson(w, sf)
	} else if r.Method == "POST" {
		var call = login1Object.Call(endpoint[sf.Self], dbus.Flags(0), false)
		if call.Err != nil {
			respond.ServerError(w, call.Err)
		} else {
			respond.Accepted(w)
		}
	} else {
		respond.NotAllowed(w)
	}
}

func Collect() respond.StandardFormatList {
	var sfl = make(respond.StandardFormatList, 0, len(actions))
	for _, action := range actions {
		var copy = *action
		sfl = append(sfl, &copy)
	}
	return sfl
}

func AllPaths() []string {
	var paths = make([]string, 0, len(actions)+1)
	for path := range actions {
		paths = append(paths, path)
	}
	paths = append(paths, "/session/actions")
	return paths
}

const managerInterface = "org.freedesktop.login1.Manager"

var dbusConn = func() *dbus.Conn {
	if conn, err := dbus.SystemBus(); err != nil {
		panic(err)
	} else {
		return conn
	}
}()

var login1Object = dbusConn.Object("org.freedesktop.login1", "/org/freedesktop/login1")

// TODO logout
var allActions = []*respond.StandardFormat{
	&respond.StandardFormat{
		Self:     "/session/shutdown",
		Type:     "session_action",
		Title:    "Shutdown",
		Comment:  "Power off the machine",
		IconName: "system-shutdown",
		OnPost:   "Shutdown",
		Data: map[string]string{
			"DbusEndpoint":          "org.freedesktop.login1.Manager.PowerOff",
			"DbusEndpointAvailable": "org.freedesktop.login1.Manager.CanPowerOff",
		},
	},
	{
		Self:     "/session/suspend",
		Type:     "session_action",
		Title:    "Suspend",
		Comment:  "Suspend the machine",
		IconName: "system-suspend",
		OnPost:   "Suspend",
		Data: map[string]string{
			"DbusEndpoint":          "org.freedesktop.login1.Manager.Suspend",
			"DbusEndpointAvailable": "org.freedesktop.login1.Manager.CanSuspend",
		},
	},
	{
		Self:     "/session/hibernate",
		Type:     "session_action",
		Title:    "Hibernate",
		Comment:  "Put the machine into hibernation",
		IconName: "system-suspend-hibernate",
		OnPost:   "Hibernate",
		Data: map[string]string{
			"DbusEndpoint":          "org.freedesktop.login1.Manager.Hibernate",
			"DbusEndpointAvailable": "org.freedesktop.login1.Manager.Hibernate",
		},
	},
	{
		Self:     "/session/hybridsleep",
		Type:     "session_action",
		Title:    "Hybrid sleep",
		Comment:  "Put the machine into hybrid sleep",
		IconName: "system-suspend-hibernate",
		OnPost:   "Hybrid sleep",
		Data: map[string]string{
			"DbusEndpoint":          "org.freedesktop.login1.Manager.HybridSleep",
			"DbusEndpointAvailable": "org.freedesktop.login1.Manager.HybridSleep",
		},
	},
	{
		Self:     "/session/reboot",
		Type:     "session_action",
		Title:    "Reboot",
		Comment:  "Reboot the machine",
		IconName: "system-reboot",
		OnPost:   "Reboot",
		Data: map[string]string{
			"DbusEndpoint":          "org.freedesktop.login1.Manager.Reboot",
			"DbusEndpointAvailable": "Reboot",
		},
	},
}

var actions = make(map[string]*respond.StandardFormat)

//

var endpoint = map[string]string{
	"/session/shutdown":    "org.freedesktop.login1.Manager.PowerOff",
	"/session/suspend":     "org.freedesktop.login1.Manager.Suspend",
	"/session/hibernate":   "org.freedesktop.login1.Manager.Hibernate",
	"/session/hybridsleep": "org.freedesktop.login1.Manager.HybridSleep",
	"/session/reboot":      "org.freedesktop.login1.Manager.Reboot",
}

var availabilityEndpoint = map[string]string{
	"/session/shutdown":    "org.freedesktop.login1.Manager.CanPowerOff",
	"/session/suspend":     "org.freedesktop.login1.Manager.CanSuspend",
	"/session/hibernate":   "org.freedesktop.login1.Manager.CanHibernate",
	"/session/hybridsleep": "org.freedesktop.login1.Manager.CanHybridSleep",
	"/session/reboot":      "org.freedesktop.login1.Manager.CanReboot",
}

func init() {
	for _, action := range allActions {
		if "yes" == login1Object.Call(availabilityEndpoint[action.Self], dbus.Flags(0)).Body[0].(string) {
			actions[action.Self] = action
		}
	}
}
