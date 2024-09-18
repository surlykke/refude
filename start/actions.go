package start

import (
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/tr"
)

type key string

const (
	shutdown key = "shutdown"
	reboot       = "reboot"
	suspend      = "suspend"
)

var actionLinks = resource.LinkList{
	{Href: "/start?action=shutdown", Title: tr.Tr("Power off"), IconUrl: "/icon?name=system-shutdown", Relation: relation.Action},
	{Href: "/start?action=reboot", Title: tr.Tr("Reboot"), IconUrl: "/icon?name=system-reboot", Relation: relation.Action},
	{Href: "/start?action=suspend", Title: tr.Tr("Suspend"), IconUrl: "/icon?name=system-suspend", Relation: relation.Action},
}

var dbusCommands = map[string]string{
	"shutdown": "org.freedesktop.login1.Manager.PowerOff",
	"reboot":   "org.freedesktop.login1.Manager.Reboot",
	"suspend":  "org.freedesktop.login1.Manager.Suspend",
}

func getStartLinks(collector *resource.LinkList) {
	*collector = append(*collector, actionLinks...)
}

func getExec(k string) ([]string, bool) {
	if dbusCommand, ok := dbusCommands[k]; ok {
		return []string{
			"dbus-send", "--system", "--print-reply", "--dest=org.freedesktop.login1", "/org/freedesktop/login1", dbusCommand, "boolean:false",
		}, true
	} else {
		return []string{}, false
	}
}
