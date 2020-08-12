package session

import (
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"

	"github.com/godbus/dbus/v5"
)

type Session struct {
	respond.Links `json:"_links"`
	// More to do
}

var session = Session{}

func Handler(r *http.Request) http.Handler {
	if r.URL.Path == "/session" {
		return session
	} else {
		return nil
	}
}

func (s Session) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, s)
	} else if r.Method == "POST" {
		var actionId = requests.GetSingleQueryParameter(r, "actionid", "suspend")
		if ep, ok := endpoint[actionId]; !ok {
			respond.NotFound(w)
		} else {
			var call = login1Object.Call(ep, dbus.Flags(0), false)
			if call.Err != nil {
				respond.ServerError(w, call.Err)
			} else {
				respond.Accepted(w)
			}
		}
	} else {
		respond.NotAllowed(w)
	}
}

func Collect() respond.Links {
	return respond.Links{session.Link()}
}

func DesktopSearch(term string, baserank int) (respond.Link, bool) {
	var rank int
	var ok bool
	for _, link := range session.Links {
		if rank, ok = searchutils.Rank(strings.ToLower(link.Title), term, baserank); ok {
			break
		}
	}
	if ok {
		var link = session.Link()
		link.Rank = rank
		return link, true
	} else {
		return respond.Link{}, false
	}
}

func AllPaths() []string {
	return []string{"/session"}
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
var allLinks = map[string]respond.Link{
	"shutdown": {
		Href:  "/session?actionid=shutdown",
		Rel:   respond.Action,
		Title: "Shutdown",
		Icon:  "system-shutdown",
	},
	"suspend": {
		Href:  "/session?actionid=suspend",
		Rel:   respond.Self,
		Title: "Suspend",
		Icon:  "system-suspend",
	},
	"hibernate": {
		Href:  "/session?actionid=hibernate",
		Rel:   respond.Action,
		Title: "Hibernate",
		Icon:  "system-suspend-hibernate",
	},
	"hybridsleep": {
		Href:  "/session?action=hybridsleep",
		Rel:   respond.Action,
		Title: "Hybrid sleep",
		Icon:  "system-suspend-hibernate",
	},
	"reboot": {
		Href:  "/session?action=reboot",
		Rel:   respond.Action,
		Title: "Reboot",
		Icon:  "system-reboot",
	},
}

var endpoint = map[string]string{
	"shutdown":    "org.freedesktop.login1.Manager.PowerOff",
	"suspend":     "org.freedesktop.login1.Manager.Suspend",
	"hibernate":   "org.freedesktop.login1.Manager.Hibernate",
	"hybridsleep": "org.freedesktop.login1.Manager.HybridSleep",
	"reboot":      "org.freedesktop.login1.Manager.Reboot",
}

var availabilityEndpoint = map[string]string{
	"shutdown":    "org.freedesktop.login1.Manager.CanPowerOff",
	"suspend":     "org.freedesktop.login1.Manager.CanSuspend",
	"hibernate":   "org.freedesktop.login1.Manager.CanHibernate",
	"hybridsleep": "org.freedesktop.login1.Manager.CanHybridSleep",
	"reboot":      "org.freedesktop.login1.Manager.CanReboot",
}

func init() {
	session.Links = make(respond.Links, 0, 5)
	session.Links = append(session.Links, allLinks["suspend"]) // Assume this is always available (?)
	for _, action := range []string{"reboot", "shutdown", "hibernate", "hybridsleep"} {
		if "yes" == login1Object.Call(availabilityEndpoint[action], dbus.Flags(0)).Body[0].(string) {
			session.Links = append(session.Links, allLinks[action])
		}
	}
}
