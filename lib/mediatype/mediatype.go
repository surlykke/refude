package mediatype

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
	Application:  "Application",
	Window:       "Window",
	Tab:          "Tab",
	File:         "File",
	Device:       "Device",
	Notification: "Notification",
	Trayitem:     "Trayitem",
	Menu:         "Menu",
	Start:        "Start",
	Mimetype:     "Mimetype",
}

func (m MediaType) Short() string {
	return short[m]
}
