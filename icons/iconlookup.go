package icons

import (
	"math"
	"os"
	"strings"
)

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
func findIconPath(name string, size uint32) string {
	lock.Lock()
	defer lock.Unlock()
	for name != "" {
		var ns = nameAndSize{name, size}
		if path := iconPathCache[ns]; path != "" {
			return path
		} else if path = locateIcon(name, size); path != "" {
			iconPathCache[ns] = path
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

func locateIcon(iconName string, size uint32) string {
	var path = ""
	for _, themeId := range themeSearchList {
		if path = locateIconInTheme(iconName, size, themeMap[themeId]); path != "" {
			break
		}
	}
	if path == "" {
		path = locateNonThemedIcon(iconName)
	}
	return path
}

func locateIconInTheme(iconName string, size uint32, theme *IconTheme) string {
	var shortestDistanceSoFar = uint32(math.MaxUint32)
	var filename string

	for _, basedir := range basedirs {
		if existsAndIsDir(basedir + "/" + theme.Id) {
			for _, iconDir := range theme.Dirs {
				if existsAndIsDir(basedir + "/" + theme.Id + "/" + iconDir.Path) {
					for _, extension := range []string{".png", ".svg"} { // TODO deal with xpm
						var tentativeFilename = basedir + "/" + theme.Id + "/" + iconDir.Path + "/" + iconName + extension
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

func locateNonThemedIcon(iconName string) string {
	for _, basedir := range basedirs {
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
