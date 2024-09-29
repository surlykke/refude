package icons

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sync"

	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type IconPath struct {
	Path    string
	MinSize uint32
	MaxSize uint32
}

var sessionIconsDir = xdg.RuntimeDir + "/org.refude.session-icons"

func init() {
	if err := os.MkdirAll(sessionIconsDir, 0700); err != nil {
		log.Warn("Unable to create ", sessionIconsDir, err)
	}
}

var themedIcons = make(map[string][]IconPath)
var sessionIcons = make(map[string][]IconPath)
var iconLock sync.Mutex

func getIconPaths(iconName string) ([]IconPath, bool) {
	iconLock.Lock()
	defer iconLock.Unlock()
	if iconPaths, ok := themedIcons[iconName]; ok {
		return iconPaths, true
	} else if iconPaths, ok = sessionIcons[iconName]; ok {
		return iconPaths, true
	} else {
		return nil, false
	}

}

func addSessionIcon(name string, iconPaths []IconPath) {
	iconLock.Lock()
	defer iconLock.Unlock()
	if _, ok := themedIcons[name]; !ok {
		sessionIcons[name] = iconPaths
	}
}

func addSessionIconSinglePath(name, path string) {
	addSessionIcon(name, []IconPath{{Path: path, MinSize: 1, MaxSize: 1}})
}

func collectIcons() {
	var collected = make(map[string][]IconPath, 500)
	collectThemeIcons(collected)
	collectOtherIcons(collected)
	iconLock.Lock()
	defer iconLock.Unlock()
	themedIcons = collected
}

/*
Somewhat based on https://specifications.freedesktop.org/icon-theme-spec/icon-theme-spec-latest.html#icon_lookup .
icon scale is ignored (TODO)

We prefer an icon from theme with not-matching size over icon from parent theme with matching size. This should
give a gui a more consistent look
*/
func collectThemeIcons(collected map[string][]IconPath) {
	var searchOrder = determineSearchOrder()
	for _, themeId := range searchOrder {
		for name, iconPaths := range collectIconsFromTheme(themeId) {
			if _, ok := collected[name]; !ok {
				collected[name] = iconPaths
			}
		}
	}
}

func collectIconsFromTheme(themeId string) map[string][]IconPath {
	var iconsFromTheme = make(map[string][]IconPath)
	var theme, _ = repo.Get[*IconTheme]("/icontheme/" + themeId)
	for _, basedir := range xdg.IconBasedirs {
		for _, themeDir := range theme.Dirs {
			var glob = basedir + "/" + themeId + "/" + themeDir.Path + "/*"
			if filePathsInThemeDir, err := filepath.Glob(glob); err == nil {
				for _, filePath := range filePathsInThemeDir {
					var ext = path.Ext(filePath)
					if ext == ".png" || ext == ".svg" { // TODO xpm
						var fileName = path.Base(filePath)
						var iconPath = IconPath{Path: filePath, MinSize: themeDir.MinSize, MaxSize: themeDir.MaxSize}
						var name = fileName[0 : len(fileName)-4]
						iconsFromTheme[name] = append(iconsFromTheme[name], iconPath)
					}
				}
			}
		}
	}
	return iconsFromTheme
}

func collectOtherIcons(collected map[string][]IconPath) {
	var dirsToLookAt = make([]string, 0, len(xdg.IconBasedirs)+1)
	dirsToLookAt = append(dirsToLookAt, xdg.IconBasedirs...)
	dirsToLookAt = append(dirsToLookAt, xdg.PixmapDir)

	for _, dir := range dirsToLookAt {
		if filePathsInDir, err := filepath.Glob(dir + "/*"); err == nil {
			for _, filePath := range filePathsInDir {
				var ext = path.Ext(filePath)
				if ext == ".png" || ext == ".svg" { // TODO xpm
					var name = path.Base(filePath)
					name = name[0 : len(name)-4]
					if _, ok := collected[name]; !ok { // We won't let a non-themed icon shadow a themed icon. cf. above.
						collected[name] = []IconPath{{Path: filePath, MinSize: 1, MaxSize: 1}}
					}
				}

			}
		}
	}
}

func determineSearchOrder() []string {
	var searchOrder = make([]string, 0, 4)
	var walker func(themeId string)
	walker = func(themeId string) {
		if themeId != "" && themeId != "hicolor" && !slices.Contains(searchOrder, themeId) {
			if theme, ok := repo.Get[*IconTheme]("/icontheme/" + themeId); ok {
				searchOrder = append(searchOrder, themeId)
				for _, parent := range theme.Inherits {
					walker(parent)
				}
			}
		}
	}
	walker(determineDefaultThemeId())
	searchOrder = append(searchOrder, "hicolor") // hicolor is the general fall back
	return searchOrder
}

func determineDefaultThemeId() string {
	var iconThemeDefPattern = regexp.MustCompile("gtk-icon-theme-name=(\\S+)")

	if defaultThemeId, ok := os.LookupEnv("REFUDE_ICON_THEME"); ok {
		return defaultThemeId
	} else {
		for _, iniFile := range []string{
			xdg.ConfigHome + "/gtk-4.0/settings.ini",
			"/etc/gtk-4.0/settings.ini",
			xdg.ConfigHome + "/gtk-3.0/settings.ini",
			"/etc/gtk-3.0/settings.ini",
			xdg.ConfigHome + "/gtk-2.0/settings.ini",
			"/etc/gtk-2.0/settings.ini",
			xdg.Home + "/.gtkrc-2.0", "/etc/gtk-2.0/gtkrc"} {

			if bytes, err := os.ReadFile(iniFile); err == nil {
				if matches := iconThemeDefPattern.FindStringSubmatch(string(bytes)); matches != nil {
					fmt.Println("Using", matches[1], "as default theme")
					return matches[1]
				}
			}
		}
	}

	return ""
}
