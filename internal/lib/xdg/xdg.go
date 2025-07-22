// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package xdg

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"github.com/surlykke/refude/internal/lib/slice"
)

var Home string
var ConfigHome string
var ConfigDirs []string
var CacheHome string
var DataHome string
var DataDirs []string
var IconBasedirs []string
var PixmapDir string
var RuntimeDir string
var CurrentDesktop []string
var Locale string
var SessionType string

var DesktopDir string
var DownloadDir string
var TemplatesDir string
var PublicshareDir string
var DocumentsDir string
var MusicDir string
var PicturesDir string
var VideosDir string

var NmSocketPath string

func init() {
	Home = clean(os.Getenv("HOME"))
	ConfigHome = clean(coalesce(os.Getenv("XDG_CONFIG_HOME"), Home+"/.config"))
	ConfigDirs = cleanS(slice.Split(coalesce(os.Getenv("XDG_CONFIG_DIRS"), "/etc/xdg"), ":"))
	CacheHome = clean(coalesce(os.Getenv("XDG_CACHE_HOME"), Home+"/.cache"))
	DataHome = clean(coalesce(os.Getenv("XDG_DATA_HOME"), Home+"/.local/share"))
	DataDirs = cleanS(slice.Split(coalesce(os.Getenv("XDG_DATA_DIRS"), "/usr/local/share:/usr/share"), ":"))
	DataDirs = slice.Remove(DataDirs, DataHome)

	IconBasedirs = []string{Home + "/.icons", DataHome + "/icons"} // Weirdly icontheme specification does not mention ~/.local/share/icons, which I consider to be an error
	for _, dataDir := range DataDirs {
		IconBasedirs = append(IconBasedirs, dataDir+"/icons")
	}
	PixmapDir = "/usr/share/pixmaps"

	RuntimeDir = clean(coalesce(os.Getenv("XDG_RUNTIME_DIR"), "/tmp"))
	CurrentDesktop = slice.Split(coalesce(os.Getenv("XDG_CURRENT_DESKTOP"), ""), ":")
	Locale = coalesce(os.Getenv("LANG"), "") // TODO Look at other env variables too

	// Strip away encoding part (ie. '.UTF-8')
	if index := strings.Index(Locale, "."); index > -1 {
		Locale = Locale[0:index]
	}
	SessionType = coalesce(os.Getenv("XDG_SESSION_TYPE"), "")

	// User dirs. Defaults taken from my /etc/xdg/user-dirs.defaults. We probably should re-read that file on startup,
	// but given that so many apps use these, I find it unlikely that they will change. (The defaults, that is)
	var userDirs, _ = readUserDirs(Home, ConfigHome)
	DesktopDir = clean(coalesce(userDirs["XDG_DESKTOP_DIR"], Home+"/Desktop"))
	DownloadDir = clean(coalesce(userDirs["XDG_DOWNLOAD_DIR"], Home+"/Download"))
	TemplatesDir = clean(coalesce(userDirs["XDG_TEMPLATES_DIR"], Home+"/Templates"))
	PublicshareDir = clean(coalesce(userDirs["XDG_PUBLICSHARE_DIR"], Home+"/public"))
	DocumentsDir = clean(coalesce(userDirs["XDG_DOCUMENTS_DIR"], Home+"/Documents"))
	MusicDir = clean(coalesce(userDirs["XDG_MUSIC_DIR"], Home+"/Music"))
	PicturesDir = clean(coalesce(userDirs["XDG_PICTURES_DIR"], Home+"/Pictures"))
	VideosDir = clean(coalesce(userDirs["XDG_VIDEOS_DIR"], Home+"/Videos"))

	NmSocketPath = RuntimeDir + "/org.refude.nm-socket"
}

func RunCmd(argv ...string) error {
	var cmd = exec.Command(argv[0], argv[1:]...)

	os.Unsetenv("LD_PRELOAD") // We don't want this passed on to launced apps
	cmd.Dir = Home
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // So ctrl-C against RefudeDesktopService doesn't affect

	if err := cmd.Start(); err == nil {
		go cmd.Wait()
		return nil
	} else {
		return err
	}
}

/*
*

	Returns the first non-empty of its arguments or "" if none found.
*/
func coalesce(args ...string) string {
	for _, arg := range args {
		if arg != "" {
			return arg
		}
	}
	return ""
}

func DirOrFileExists(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func clean(dir string) string {
	return path.Clean(dir)
}

func cleanS(dirs []string) []string {
	var cleaned = make([]string, 0, len(dirs))
	for _, dir := range dirs {
		cleaned = append(cleaned, path.Clean(dir))
	}
	return cleaned
}
