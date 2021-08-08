package icons

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

var (
	lock             sync.Mutex
	basedirs         = []string{}
	addedBasedirs    = make(map[string]bool)
	themeMap         ThemeMap
	defaultThemeName string
	themeList        = make([]*IconTheme, 0, 3) // First, defaultTheme, if given, then those directly or indirectly inherited
	hicolor          *IconTheme
	other            = make(map[string]IconImage)
)

func init() {
	determineBasedirs()
	themeMap = readThemes()
	determineDefaultIconTheme()
	addInheritedThemesToThemeList()
	for _, theme := range themeList {
		for _, basedir := range basedirs {
			collectThemeIcons(theme, basedir)
		}
	}

	if hicolor = themeMap["hicolor"]; hicolor != nil {
		for _, basedir := range basedirs {
			collectThemeIcons(hicolor, basedir)
		}
	}

	for _, basedir := range basedirs {
		collecOtherIcons(basedir)
	}
}

func determineBasedirs() {
	basedirs = append(basedirs, xdg.Home+"/.icons")
	basedirs = append(basedirs, xdg.DataHome+"/icons") // Weirdly icontheme specification does not mention this, which I consider to be an error
	for _, dataDir := range xdg.DataDirs {
		basedirs = append(basedirs, dataDir+"/icons")
	}
	basedirs = append(basedirs, "/usr/share/pixmaps")
}

/**
 *Finds, if possible, the default theme, and places it first in themeList
 */
func determineDefaultIconTheme() {
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

	if defaultThemeName, ok = os.LookupEnv("REFUDE_ICON_THEME"); !ok {
		for _, fileToLookAt := range filesToLookAt {
			if bytes, err := ioutil.ReadFile(fileToLookAt); err == nil {
				if matches := iconThemeDefPattern.FindStringSubmatch(string(bytes)); matches != nil {
					defaultThemeName = matches[1]
					break
				}
			}
		}
	}

	if defaultThemeName != "" {
		for _, theme := range themeMap {
			if theme.Name == defaultThemeName {
				themeList = []*IconTheme{theme}
				return
			}
		}
	}
}

func addInheritedThemesToThemeList() {
	var themeIsAddedOrIsHicolor = func(themeId string) bool {
		if themeId == "hicolor" {
			return true
		} else {
			for _, addedTheme := range themeList {
				if themeId == addedTheme.Id {
					return true
				}
			}
		}
		return false
	}

	for i := 0; i < len(themeList); i++ {
		for _, inheritedId := range themeList[i].Inherits {
			if !themeIsAddedOrIsHicolor(inheritedId) {
				if theme, ok := themeMap[inheritedId]; !ok {
					log.Info("Don't have theme", inheritedId)
				} else {
					themeList = append(themeList, theme)
				}
			}
		}
	}
}

func collectThemeIcons(it *IconTheme, basedir string) {
	lock.Lock()
	defer lock.Unlock()

	if dirExists(basedir + "/" + it.Id) {
		for _, dir := range it.Dirs {
			var filePattern = basedir + "/" + it.Id + "/" + dir.Path + "/*"
			if matches, err := filepath.Glob(filePattern); err == nil {
				for _, match := range matches {
					if iconName, data, ok := iconNameAndData(match); ok {
						icon, ok := it.icons[iconName]
						if !ok {
							icon = &Icon{Name: iconName, Theme: it.Name}
							it.icons[iconName] = icon
						}
						addImageToIcon(icon, IconImage{Context: dir.Context, MinSize: dir.MinSize, MaxSize: dir.MaxSize, Path: match, Data: data})
					}
				}
			} else {
				log.Warn("Problem with search:", filePattern, err)
			}

		}
	}
}

func collecOtherIcons(basedir string) {
	lock.Lock()
	defer lock.Unlock()

	if matches, err := filepath.Glob(basedir + "/*"); err == nil {
		for _, match := range matches {
			if iconName, data, ok := iconNameAndData(match); ok {
				if _, ok := other[iconName]; !ok {
					other[iconName] = IconImage{Path: match, Data: data}
				}
			}
		}
	} else {
		log.Warn("Could not match", basedir+"/*", err)
	}
}

/**
* returns iconName if path points to an icon.
* If the icon is of type xpm, it is converted to png, and the bytes are returned as 2. return val, which is nil otherwise
 */
func iconNameAndData(path string) (string, []byte, bool) {
	if fileInfo, err := os.Stat(path); err != nil {
		log.Warn("Could not handle", path, err)
		return "", nil, false
	} else if fileInfo.IsDir() {
		return "", nil, false
	} else if fileInfo.Mode()&(os.ModeDevice|os.ModeNamedPipe|os.ModeSocket|os.ModeCharDevice) != 0 {
		return "", nil, false
	} else if !(strings.HasSuffix(fileInfo.Name(), ".png") || strings.HasSuffix(fileInfo.Name(), ".svg") || strings.HasSuffix(fileInfo.Name(), ".xpm")) {
		return "", nil, false
	} else {
		var iconName = fileInfo.Name()[:len(fileInfo.Name())-4]
		var iconType = fileInfo.Name()[len(fileInfo.Name())-3:]

		if iconType == "xpm" {
			if xpmBytes, err := ioutil.ReadFile(path); err != nil {
				return "", nil, false
			} else {
				return iconName, xpmBytes, true
			}
		} else {
			return iconName, nil, true
		}
	}
}

func addImageToIcon(icon *Icon, image IconImage) {
	icon.Images = append(icon.Images, image) // TODO sort by size, somehow
}

func AddX11Icon(data []uint32) (string, error) {
	var iconName = image.X11IconHashName(data)

	lock.Lock()
	defer lock.Unlock()

	if _, ok := hicolor.icons[iconName]; !ok {
		if pngList, err := image.X11IconToPngs(data); err != nil {
			log.Warn("Error converting:", err)
			return "", err
		} else {
			icon := &Icon{Name: iconName, Theme: "Hicolor", Images: make([]IconImage, 0, len(pngList))}
			for _, sizedPng := range pngList {
				if sizedPng.Width != sizedPng.Height {
					log.Warn("Ignore image", sizedPng.Width, "x", sizedPng.Height, ", not square")
				} else {
					icon.Images = append(icon.Images, IconImage{MinSize: sizedPng.Width, MaxSize: sizedPng.Width, Data: sizedPng.Data})
				}
			}

			if len(icon.Images) == 0 {
				return "", fmt.Errorf("No usable images")
			} else {
				hicolor.icons[iconName] = icon
			}
		}
	}
	return iconName, nil
}

func AddARGBIcon(argbIcon image.ARGBIcon) string {
	var iconName = image.ARGBIconHashName(argbIcon)
	lock.Lock()
	defer lock.Unlock()

	if _, ok := hicolor.icons[iconName]; !ok {
		var icon = &Icon{Name: iconName, Theme: "Hicolor", Images: make([]IconImage, 0, len(argbIcon.Images))}
		for _, pixMap := range argbIcon.Images {
			if pixMap.Width == pixMap.Height { // else ignore
				if png, err := pixMap.AsPng(); err != nil {
					log.Warn("Unable to convert image", err)
				} else {
					icon.Images = append(icon.Images, IconImage{MinSize: pixMap.Width, MaxSize: pixMap.Width, Data: png})
				}
			}
		}
		if len(icon.Images) == 0 {
			return ""
		} else {
			hicolor.icons[iconName] = icon
		}
	}
	return iconName
}

func AddFileIcon(filePath string) (string, error) {
	filePath = path.Clean(filePath)
	var name = strings.Replace(filePath[1:len(filePath)-4], "/", ".", -1)
	lock.Lock()
	defer lock.Unlock()
	if _, ok := other[name]; !ok {

		if fileInfo, err := os.Stat(filePath); err != nil {
			return "", err
		} else if !fileInfo.Mode().IsRegular() {
			return "", fmt.Errorf("Not a regular file: %s", filePath)
		} else if !(strings.HasSuffix(filePath, ".png") || strings.HasSuffix(filePath, ".svg")) {
			return "", fmt.Errorf("Not an icon  file %s", filePath)
		} else {
			other[name] = IconImage{Path: filePath}
		}
	}
	return name, nil
}

func AddRawImageIcon(imageData image.ImageData) string {
	iconName := image.ImageDataHashName(imageData)
	lock.Lock()
	defer lock.Unlock()
	if _, ok := other[iconName]; !ok {
		if png, err := imageData.AsPng(); err != nil {
			log.Warn("Error converting image", err)
			return ""
		} else {
			other[iconName] = IconImage{Data: png}
		}
	}
	return iconName
}

func AddBasedir(basedir string) {
	if noteDirAsAdded(basedir) {
		return
	}

	for _, theme := range themeList {
		collectThemeIcons(theme, basedir)
	}
	collectThemeIcons(hicolor, basedir)
	collecOtherIcons(basedir)
}

// Returns whether basedir was already added
func noteDirAsAdded(basedir string) bool {
	lock.Lock()
	defer lock.Unlock()
	var hasBeenAdded = addedBasedirs[basedir]
	addedBasedirs[basedir] = true
	return hasBeenAdded
}

func dirExists(path string) bool {
	var info, err = os.Stat(path)
	return err == nil && info.IsDir()
}

func Run() {}
