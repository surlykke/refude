package mediatype

import "github.com/surlykke/RefudeServices/lib/translate"

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
	IconTheme    MediaType = "application/vnd.org.refude.icontheme+json"
	Bookmark     MediaType = "application/vnd.org.refude.bookmark+json"
)

var short = map[MediaType]string{
	Application:  translate.Text("Application"),
	Window:       translate.Text("Window"),
	Tab:          translate.Text("Tab"),
	File:         translate.Text("File"),
	Device:       translate.Text("Device"),
	Notification: translate.Text("Notification"),
	Trayitem:     translate.Text("Trayitem"),
	Menu:         translate.Text("Menu"),
	Start:        translate.Text("Start"),
	Mimetype:     translate.Text("Mimetype"),
	Bookmark:     translate.Text("Bookmark"),
}

func (m MediaType) Short() string {
	return short[m]
}
