package mediatype

import "github.com/surlykke/RefudeServices/lib/tr"

type MediaType string

const (
	Application  MediaType = "application/vnd.org.refude.application+json"
	Window       MediaType = "application/vnd.org.refude.window+json"
	Tab          MediaType = "application/vnd.org.refude.tab+json"
	File         MediaType = "application/vnd.org.refude.file+json"
	Device       MediaType = "application/vnd.org.refude.device+json"
	Notification MediaType = "application/vnd.org.refude.notification+json"
	Trayitem     MediaType = "application/vnd.org.refude.trayitem+json"
	Menu         MediaType = "application/vnd.org.refude.menu+json"
	Start        MediaType = "application/vnd.org.refude.start+json"
	Mimetype     MediaType = "application/vnd.org.refude.mimetype+json"
)

var short = map[MediaType]string{
	Application:  tr.Tr("Application"),
	Window:       tr.Tr("Window"),
	Tab:          tr.Tr("Tab"),
	File:         tr.Tr("File"),
	Device:       tr.Tr("Device"),
	Notification: tr.Tr("Notification"),
	Trayitem:     tr.Tr("Trayitem"),
	Menu:         tr.Tr("Menu"),
	Start:        tr.Tr("Start"),
	Mimetype:     tr.Tr("Mimetype"),
}

func (m MediaType) Short() string {
	return short[m]
}
