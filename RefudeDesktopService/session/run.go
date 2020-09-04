package session

import (
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"
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
		var actionId = requests.GetSingleQueryParameter(r, "actionid", "")
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

func DesktopSearch(term string, baserank int) respond.Links {
	var rank int
	var ok bool
	var result = make(respond.Links, 0, 6)

	for _, link := range session.Links {
		var rel respond.Relation = respond.Action
		if link.Rel == respond.Self {
			rel = respond.Related
		}
		if rank, ok = searchutils.Rank(strings.ToLower(link.Title), term, baserank); ok {
			result = append(result, respond.Link{
				Href:  link.Href,
				Rel:   rel,
				Title: link.Title,
				Icon:  link.Icon,
				Rank:  rank,
			})
		}
	}
	return result
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
		Icon:  icons.IconUrl("system-shutdown"),
	},
	"suspend": {
		Href:  "/session?actionid=suspend",
		Rel:   respond.Action,
		Title: "Suspend",
		Icon:  icons.IconUrl("system-suspend"),
	},
	"hibernate": {
		Href:  "/session?actionid=hibernate",
		Rel:   respond.Action,
		Title: "Hibernate",
		Icon:  icons.IconUrl("system-suspend-hibernate"),
	},
	"hybridsleep": {
		Href:  "/session?action=hybridsleep",
		Rel:   respond.Action,
		Title: "Hybrid sleep",
		Icon:  icons.IconUrl("system-suspend-hibernate"),
	},
	"reboot": {
		Href:  "/session?action=reboot",
		Rel:   respond.Action,
		Title: "Reboot",
		Icon:  icons.IconUrl("system-reboot"),
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
	session.Links = make(respond.Links, 0, 6)
	session.Links = append(session.Links, respond.Link{
		Href:    "/session",
		Rel:     respond.Self,
		Title:   "Session",
		Profile: "/profile/session",
	})
	for _, action := range []string{"suspend", "reboot", "shutdown", "hibernate", "hybridsleep"} {
		if "yes" == login1Object.Call(availabilityEndpoint[action], dbus.Flags(0)).Body[0].(string) {
			session.Links = append(session.Links, allLinks[action])
		}
	}
}
