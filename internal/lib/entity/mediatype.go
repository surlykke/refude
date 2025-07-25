// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package entity

import (
	"encoding/json"

	"github.com/surlykke/refude/internal/lib/translate"
)

type MediaType uint8

const (
	Application MediaType = iota
	Window
	Tab
	File
	Device
	Notification
	Trayitem
	Menu
	Start
	Mimetype
	IconTheme
	Bookmark
)

func (m MediaType) String() string {
	switch m {
	case Application:
		return "application/vnd.org.refude.application+json"
	case Window:
		return "application/vnd.org.refude.window+json"
	case Tab:
		return "application/vnd.org.refude.tab+json"
	case File:
		return "application/vnd.org.refude.file+json"
	case Device:
		return "application/vnd.org.refude.device+json"
	case Notification:
		return "application/vnd.org.refude.notification+json"
	case Trayitem:
		return "application/vnd.org.refude.trayitem+json"
	case Menu:
		return "application/vnd.org.refude.menu+json"
	case Start:
		return "application/vnd.org.refude.start+json"
	case Mimetype:
		return "application/vnd.org.refude.mimetype+json"
	case IconTheme:
		return "application/vnd.org.refude.icontheme+json"
	case Bookmark:
		return "application/vnd.org.refude.bookmark+json"
	default:
		return ""
	}
}

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
	Bookmark:     "Bookmark",
}

func (m *MediaType) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

func (m MediaType) Short() string {
	return translate.Text(short[m])
}

func (m MediaType) ShortPlural() string {
	return translate.Text(short[m] + "s")
}
