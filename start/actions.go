package start

import (
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type key string

const (
	shutdown key = "shutdown"
	reboot       = "reboot"
	suspend      = "suspend"
)

var dbusCommands = map[key]string{
	shutdown: "PowerOff",
	reboot:   "Reboot",
	suspend:  "Suspend",
}

var localizedTitles = map[key]map[string]string{
	shutdown: {
		"":   "Power off",
		"da": "Sluk",
	},
	reboot: {
		"":   "Reboot",
		"da": "Genstart",
	},
	suspend: {
		"":   "Suspend",
		"da": "Slumre",
	},
}

func getStartLinks(collector *resource.LinkList) {
	for _, k := range []key{shutdown, reboot, suspend} {
		*collector = append(*collector, resource.Link{
			Href:     "/start?action=" + string(k),
			Title:    xdg.GetFromLocalizedMap(localizedTitles[k]),
			IconUrl:  "/icon?name=system-" + string(k),
			Relation: relation.Action,
		})
	}
}

func getExec(k string) ([]string, bool) {
	if dbusCommand, ok := dbusCommands[key(k)]; ok {
		return []string{
			"dbus-send", "--system", "--print-reply", "--dest=org.freedesktop.login1", "/org/freedesktop/login1", "org.freedesktop.login1.Manager." + dbusCommand, "boolean:false",
		}, true
	} else {
		return []string{}, false
	}
}
