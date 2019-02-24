// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package xdg

import (
	"fmt"
	"github.com/surlykke/RefudeServices/lib/slice"
	"log"
	"os"
	"os/exec"
	"strings"
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
var RefudeConvertedIconsDir string
var RefudeSessionIconsDir string

func init() {
	Home = os.Getenv("HOME")
	ConfigHome = notEmptyOr(os.Getenv("XDG_CONFIG_HOME"), Home+"/.config")
	ConfigDirs = slice.Split(notEmptyOr(os.Getenv("XDG_CONFIG_DIRS"), "/etc/xdg"), ":")
	CacheHome = notEmptyOr(os.Getenv("XDG_CACHE_HOME"), Home+"/.cache")
	DataHome = notEmptyOr(os.Getenv("XDG_DATA_HOME"), Home+"/.local/share")
	DataDirs = slice.Split(notEmptyOr(os.Getenv("XDG_DATA_DIRS"), "/usr/share:/usr/local/share"), ":")
	DataDirs = slice.Remove(DataDirs, DataHome)
	RuntimeDir = notEmptyOr(os.Getenv("XDG_RUNTIME_DIR"), "/tmp")
	CurrentDesktop = slice.Split(notEmptyOr(os.Getenv("XDG_CURRENT_DESKTOP"), ""), ":")
	Locale = notEmptyOr(os.Getenv("LANG"), "") // TODO Look at other env variables too
	if index := strings.Index(Locale, "."); index > -1 { // Strip away encoding part (ie. '.UTF-8')
		Locale = Locale[0:index]
	}
	RefudeConvertedIconsDir = RuntimeDir + "/org.refude.converted-icons"
	RefudeSessionIconsDir = RuntimeDir + "/org.refude.session-icons"

	if err := os.MkdirAll(RefudeConvertedIconsDir, 0700); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(RefudeConvertedIconsDir, 0700); err != nil {
		panic(err)
	}
}


func RunCmd(argv []string) {
	fmt.Println("runCmd")
	for i := 0; i < len(argv); i++ {
		fmt.Println(i, ":", argv[i])
	}
	var cmd = exec.Command(argv[0], argv[1:]...)

	cmd.Dir = Home
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		log.Println(err.Error())
		return
	}

	go cmd.Wait() // TODO Transfer parenthood to proc 1
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