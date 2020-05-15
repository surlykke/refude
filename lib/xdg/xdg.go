// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package xdg

import (
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/surlykke/RefudeServices/lib/slice"
)

var Home string
var ConfigHome string
var ConfigDirs []string
var CacheHome string
var DataHome string
var DataDirs []string
var RuntimeDir string
var CurrentDesktop []string
var Locale string

var DesktopDir string
var DownloadDir string
var TemplatesDir string
var PublicshareDir string
var DocumentsDir string
var MusicDir string
var PicturesDir string
var VideosDir string

func init() {
	Home = os.Getenv("HOME")
	ConfigHome = notEmptyOr(os.Getenv("XDG_CONFIG_HOME"), Home+"/.config")
	ConfigDirs = slice.Split(notEmptyOr(os.Getenv("XDG_CONFIG_DIRS"), "/etc/xdg"), ":")
	CacheHome = notEmptyOr(os.Getenv("XDG_CACHE_HOME"), Home+"/.cache")
	DataHome = notEmptyOr(os.Getenv("XDG_DATA_HOME"), Home+"/.local/share")
	DataDirs = slice.Split(notEmptyOr(os.Getenv("XDG_DATA_DIRS"), "/usr/local/share:/usr/share"), ":")
	DataDirs = slice.Remove(DataDirs, DataHome)
	RuntimeDir = notEmptyOr(os.Getenv("XDG_RUNTIME_DIR"), "/tmp")
	CurrentDesktop = slice.Split(notEmptyOr(os.Getenv("XDG_CURRENT_DESKTOP"), ""), ":")
	Locale = notEmptyOr(os.Getenv("LANG"), "") // TODO Look at other env variables too
	// Strip away encoding part (ie. '.UTF-8')
	if index := strings.Index(Locale, "."); index > -1 {
		Locale = Locale[0:index]
	}

	// User dirs. Defaults taken from my /etc/xdg/user-dirs.defaults. We probably should re-read that file on startup,
	// but given that so many apps use these, I find it unlikely that they will change. (The defaults, that is)
	var userDirs, _ = readUserDirs(Home, ConfigHome)
	DesktopDir = notEmptyOr(userDirs["XDG_DESKTOP_DIR"], Home+"/DESKTOP")
	DownloadDir = notEmptyOr(userDirs["XDG_DOWNLOAD_DIR"], Home+"/DOWNLOAD")
	TemplatesDir = notEmptyOr(userDirs["XDG_TEMPLATES_DIR"], Home+"/TEMPLATES")
	PublicshareDir = notEmptyOr(userDirs["XDG_PUBLICSHARE_DIR"], Home+"/PUBLICSHARE")
	DocumentsDir = notEmptyOr(userDirs["XDG_DOCUMENTS_DIR"], Home+"/DOCUMENTS")
	MusicDir = notEmptyOr(userDirs["XDG_MUSIC_DIR"], Home+"/MUSIC")
	PicturesDir = notEmptyOr(userDirs["XDG_PICTURES_DIR"], Home+"/PICTURES")
	VideosDir = notEmptyOr(userDirs["XDG_VIDEOS_DIR"], Home+"/VIDEOS")
}

func RunCmd(argv ...string) error {
	var cmd = exec.Command(argv[0], argv[1:]...)

	cmd.Dir = Home
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // So ctrl-C against RefudeDesktopService doesn't affect

	return cmd.Start()
}

func notEmptyOr(primary string, secondary string) string {
	if primary != "" {
		return primary
	} else {
		return secondary
	}
}

func DirOrFileExists(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}
