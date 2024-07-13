package icons

import (
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"

	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

// Describes what themes are in play for user
type iconThemeCollection struct {
	basedirs              []string
	themeSearchList       []*IconTheme //  themes to search. First, defaultTheme, if given, then those directly or indirectly inherited, ending with hicolor
	allThemes             map[string]*IconTheme
	iconPathCache         map[nameAndSize]string
	refudeSessionIconsDir string
}

func buildIconThemeCollection() (*iconThemeCollection, error) {
	var ic = iconThemeCollection{}
	ic.refudeSessionIconsDir = xdg.RuntimeDir + "/org.refude.session-icons"
	if err := os.MkdirAll(ic.refudeSessionIconsDir, 0700); err != nil {
		return nil, err
	}

	// ------------ determine basedirs --------------------

	ic.basedirs = make([]string, 0, 10)
	ic.basedirs = append(ic.basedirs, xdg.Home+"/.icons")
	ic.basedirs = append(ic.basedirs, xdg.DataHome+"/icons") // Weirdly icontheme specification does not mention this, which I consider to be an error
	for _, dataDir := range xdg.DataDirs {
		ic.basedirs = append(ic.basedirs, dataDir+"/icons")
	}
	ic.basedirs = append(ic.basedirs, "/usr/share/pixmaps")
	var pos = 0
	for i := 0; i < len(ic.basedirs); i++ {
		if existsAndIsDir(ic.basedirs[i]) {
			ic.basedirs[pos] = ic.basedirs[i]
			pos++
		}
	}
	ic.basedirs = ic.basedirs[0:pos]

	ic.allThemes = readThemes(ic.basedirs)
	hicolor, ok := ic.allThemes["hicolor"]
	if !ok {
		return nil, errors.New("No hicolor theme")
	}
	var sessionHicolorPath = ic.refudeSessionIconsDir + "/hicolor/"
	for _, dir := range hicolor.Dirs {
		os.MkdirAll(sessionHicolorPath+dir.Path, 0700)
	}
	var defaultThemeId = determineDefaultThemeId(ic.allThemes)
	ic.themeSearchList = buildSearchList(defaultThemeId, ic.allThemes, hicolor)
	return &ic, nil
}

func determineDefaultThemeId(themes map[string]*IconTheme) string {
	fmt.Println("determineDefaultThemeId")
	var (
		defaultThemeName = ""
		ok               bool
		filesToLookAt    = []string{
			xdg.ConfigHome + "/gtk-4.0/settings.ini",
			"/etc/gtk-4.0/settings.ini",
			xdg.ConfigHome + "/gtk-3.0/settings.ini",
			"/etc/gtk-3.0/settings.ini",
			xdg.ConfigHome + "/gtk-2.0/settings.ini",
			"/etc/gtk-2.0/settings.ini",
			xdg.Home + "/.gtkrc-2.0",
			"/etc/gtk-2.0/gtkrc",
		}

		iconThemeDefPattern = regexp.MustCompile("gtk-icon-theme-name=(\\S+)")
	)

	fmt.Println("Look in env")
	if defaultThemeName, ok = os.LookupEnv("REFUDE_ICON_THEME"); ok {
		log.Info("default icon theme taken from env REFUDE_ICON_THEME", defaultThemeName)
	} else {
		for _, fileToLookAt := range filesToLookAt {
			fmt.Println("Looking at", fileToLookAt)
			if bytes, err := os.ReadFile(fileToLookAt); err == nil {
				if matches := iconThemeDefPattern.FindStringSubmatch(string(bytes)); matches != nil {
					defaultThemeName = matches[1]
					log.Info("default icon theme taken from", fileToLookAt, defaultThemeName)
					break
				}
			}
		}
	}

	if defaultThemeName != "" {
		for _, theme := range themes {
			if theme.Title == defaultThemeName {
				return defaultThemeName
			}
		}
	}
	return ""
}

func buildSearchList(defaultThemePath string, themes map[string]*IconTheme, hicolor *IconTheme) []*IconTheme {
	var searchList []*IconTheme
	var added = map[string]bool{"hicolor": true}

	var walk func(string)
	walk = func(themeId string) { // Consider: depth first or width first...
		if !added[themeId] {
			if theme, ok := themes[themeId]; ok {
				searchList = append(searchList, theme)
				added[themeId] = true
				for _, inheritedThemeId := range theme.Inherits {
					walk(inheritedThemeId)
				}
			}
		}
	}

	walk(defaultThemePath)

	searchList = append(searchList, hicolor)
	return searchList
}

func (ic *iconThemeCollection) hicolor() *IconTheme {
	return ic.themeSearchList[len(ic.themeSearchList)-1]
}

func (ic *iconThemeCollection) writeSessionHicolorIcon(iconName string, size uint32, png []byte) {
	for _, dir := range ic.hicolor().Dirs {
		if dir.Context == "converted" && dir.MinSize <= size && dir.MaxSize >= size {
			var path = fmt.Sprintf("%s/hicolor/%s/%s.png", ic.refudeSessionIconsDir, dir.Path, iconName)
			if err := os.WriteFile(path, png, 0700); err != nil {
				log.Warn("Problem writing", path, err)
			}
			return
		}
	}

	log.Warn("Found no suitable converted dir for", iconName, "of size", size)
}

func (ic *iconThemeCollection) writeSessionOtherIcon(iconName string, data []byte) {
	var path = fmt.Sprintf("%s/%s.png", ic.refudeSessionIconsDir, iconName)
	if err := os.WriteFile(path, data, 0700); err != nil {
		log.Warn("Problem writing", path, err)
	}
}

/*
Somewhat based on https://specifications.freedesktop.org/icon-theme-spec/icon-theme-spec-latest.html#icon_lookup .
icon scale is ignored (TODO)

We prefer an icon from theme with not-matching size over icon from parent theme with matching size. This should
give a gui a more consistent look

By the icon naming specification, dash ('-') seperates 'levels of specificity'. So given an icon name
'input-mouse-usb', the levels of specificy, and the names and order we search will be: 'input-mouse-usb',
'input-mouse' and 'input'. Here we prefer specificy over theme, ie. if 'input-mouse-usb' is found in an inherited theme, that
is preferred over 'input-mouse' in the default theme
*/
func (itc *iconThemeCollection) findIconPath(name string, size uint32) string {
	for name != "" {
		var ns = nameAndSize{name, size}
		if path := itc.iconPathCache[ns]; path != "" {
			return path
		} else if path = itc.locateIcon(name, size); path != "" {
			itc.iconPathCache[ns] = path
			return path
		}

		if lastDash := strings.LastIndex(name, "-"); lastDash > -1 {
			name = name[0:lastDash]
		} else {
			name = ""
		}

	}
	return ""
}

func (itc *iconThemeCollection) locateIcon(iconName string, size uint32) string {
	var path = ""
	for _, theme := range itc.themeSearchList {
		if path = itc.locateIconInTheme(iconName, size, theme); path != "" {
			break
		}
	}
	if path == "" {
		path = itc.locateNonThemedIcon(iconName)
	}
	return path
}

func (itc *iconThemeCollection) locateIconInTheme(iconName string, size uint32, theme *IconTheme) string {
	var shortestDistanceSoFar = uint32(math.MaxUint32)
	var filename string
	var id = theme.Path[len("/icontheme/"):]

	for _, basedir := range itc.basedirs {
		if existsAndIsDir(basedir + "/" + id) {
			for _, iconDir := range theme.Dirs {
				if existsAndIsDir(basedir + "/" + id + "/" + iconDir.Path) {
					for _, extension := range []string{".png", ".svg"} { // TODO deal with xpm
						var tentativeFilename = basedir + "/" + id + "/" + iconDir.Path + "/" + iconName + extension
						if existsAndIsNotDir(tentativeFilename) /* Maybe also check mimetype... */ {
							var distance uint32
							if iconDir.MinSize > size {
								distance = iconDir.MinSize - size
							} else if iconDir.MaxSize < size {
								distance = size - iconDir.MaxSize
							} else {
								distance = 0
							}

							if distance < shortestDistanceSoFar {
								shortestDistanceSoFar = distance
								filename = tentativeFilename
								shortestDistanceSoFar = distance
							}
							if shortestDistanceSoFar == 0 {
								return filename
							}
						}
					}
				}
			}
		}
	}
	return filename
}

func (itc *iconThemeCollection) locateNonThemedIcon(iconName string) string {
	for _, basedir := range itc.basedirs {
		for _, extension := range []string{".png", ".svg"} {
			var tentativeFileName = basedir + "/" + iconName + extension
			if existsAndIsNotDir(tentativeFileName) {
				return tentativeFileName
			}
		}
	}
	return ""
}

func existsAndIsDir(path string) bool {
	var info, err = os.Stat(path)
	return err == nil && info.IsDir()
}

func existsAndIsNotDir(filepath string) bool {
	fileinfo, err := os.Stat(filepath)
	return err == nil && !fileinfo.Mode().IsDir()
}
