/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package xdg

import (
    "os"
    "strings"
)

func Home() string {
    return xdg.Home
}

func ConfigHome() string {
    return xdg.ConfigHome
}

func ConfigDirs() []string {
    return xdg.ConfigDirs
}

func CacheHome() string {
    return xdg.CacheHome
}

func DataHome() string {
    return xdg.DataHome
}

func DataDirs() []string {
    return xdg.DataDirs
}

func RuntimeDir() string {
    return xdg.RuntimeDir
}

func init() {
    xdg.Home = os.Getenv("HOME")
    xdg.ConfigHome = notEmptyOr(os.Getenv("XDG_CONFIG_HOME"), xdg.Home + "/.config")
    xdg.ConfigDirs = strings.Split(notEmptyOr(os.Getenv("XDG_CONFIG_DIRS"), "/etc/xdg"), ":")
    xdg.CacheHome = notEmptyOr(os.Getenv("XDG_CACHE_HOME"), xdg.Home + "/.cache")
    xdg.DataHome = notEmptyOr(os.Getenv("XDG_DATA_HOME"), xdg.Home + "/.local/share")
    tmp := strings.Split(notEmptyOr(os.Getenv("XDG_DATA_DIRS"), "/usr/share:/usr/local/share"), ":")
    xdg.DataDirs = make([]string, 0, len(tmp))
    for _, dataDir := range tmp {
        if dataDir != xdg.DataHome {
            xdg.DataDirs = append(xdg.DataDirs, dataDir)
        }
    }
    xdg.RuntimeDir = notEmptyOr(os.Getenv("XDG_RUNTIME_DIR"), "/tmp")
}

var xdg struct {
    Home string
    ConfigHome string
    ConfigDirs []string
    CacheHome string
    DataHome string
    DataDirs []string
    RuntimeDir string
}



func notEmptyOr(primary string, secondary string) string {
    if primary != "" {
        return primary
    } else {
        return secondary
    }
}
