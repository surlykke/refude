package main

import (
	"github.com/surlykke/RefudeServices/common"
	"github.com/surlykke/RefudeServices/xdg"
	"path/filepath"
	"fmt"
	"os"
	"math"
	"strconv"
)

type Themes struct {
	themes          map[string]Theme
	fallbackIcons   map[string][]Icon
}

type Theme struct {
	Id          string
	Name        string
	Comment     string
    Context     string
	Inherits    []string
	Ancestors   []string
	IconDirs    []IconDir

	// Maps an icon name to a list of all found icons with that name
	Icons       map[string][]Icon
}

type IconDir struct {
	Path      string
	MinSize uint32
	MaxSize uint32
	Context string
}


type Icon struct {
	Name     string
	Context  string
	MinSize  uint32
	MaxSize  uint32
	Path     string
}

var searchDirectories []string = getSearchDirectories()

func ReadThemes() Themes {
	res := Themes{make(map[string]Theme), make(map[string][]Icon)}

	for _,searcDir := range searchDirectories {
		indexThemeFilePaths, err := filepath.Glob(searcDir + "/" + "*"  + "/index.theme")
		if err != nil { panic(err) }

		for _, indexThemeFilePath := range indexThemeFilePaths {
			themeId := filepath.Base(filepath.Dir(indexThemeFilePath))
			if _,ok := res.themes[themeId]; !ok {
				if theme, err := readIndexTheme(themeId, indexThemeFilePath); err == nil {
					res.themes[themeId] = theme
				}
			}
		}
	}

	for themeId, theme := range res.themes {
		fmt.Println("parents for ", themeId, ": ", theme.Inherits)
		ancestors := getAncestors(themeId, make(common.StringSet), res.themes)
		fmt.Println("ancestors for ", themeId, ": ", ancestors)
		ancestors = append(ancestors, "hicolor")
		theme.Ancestors = ancestors
		res.themes[themeId] = theme
	}

	for _, searchDir := range searchDirectories {
		for _, theme := range res.themes {
			for _, iconDir := range theme.IconDirs {
				iconDirPath := searchDir + "/" + theme.Id + "/" + iconDir.Path
				if _, err := os.Stat(iconDirPath); err != nil {
					continue
				}
				collectIcons(theme.Icons, iconDirPath, iconDir)
			}
		}

	}

	for _, searchDir := range searchDirectories {
		dummyIconDir := IconDir{}
		collectIcons(res.fallbackIcons, searchDir, dummyIconDir)

	}

	fmt.Println("Themes read")
	for themeId, theme := range res.themes {
		fmt.Println(themeId, ": ", theme.Ancestors)
	}

	iconCount := 0
	maxListLength := 0

	for _,theme := range res.themes {
		for _,iconList := range theme.Icons {
			iconCount++
			if len(iconList) > maxListLength {
				maxListLength = len(iconList)
			}
		}
	}

	fmt.Println(iconCount, " icons, maxListLength: ", maxListLength)

	return res
}


func readIndexTheme(themeId string, indexThemeFilePath string) (Theme, error) {
	inigroups,err:= common.ReadIniFile(indexThemeFilePath)
	if err != nil {
		return Theme{}, err
	}

	if len(inigroups) < 1 || inigroups[0].Name != "Icon Theme" {
		return Theme{}, fmt.Errorf("Error reading %s , expected 'Icon Theme' at start", indexThemeFilePath)
	}

	theme := Theme{}
	theme.Id = themeId
	theme.Name = inigroups[0].Entry["Name"]
	theme.Comment = inigroups[0].Entry["Comment"]
	theme.Inherits = common.Split(inigroups[0].Entry["Inherits"], ",")
	theme.IconDirs = []IconDir{}
	directories := common.ToSet(common.Split(inigroups[0].Entry["Directories"], ","))

	for _,iniGroup := range inigroups[1:] {

		if ! directories[iniGroup.Name] {
			fmt.Fprintln(os.Stderr, iniGroup.Name, " not found in Directories")
			continue
		}

		size, sizeGiven := readUint32(iniGroup.Entry["Size"])
		if !sizeGiven  {
			fmt.Fprintln(os.Stderr, "Skipping ", iniGroup.Name, " - no size given")
			continue
		}

		minSize, minSizeGiven := readUint32(iniGroup.Entry["MinSize"])
		maxSize, maxSizeGiven := readUint32(iniGroup.Entry["MaxSize"])
		threshold := readUint32OrFallback(iniGroup.Entry["Threshold"], 2)
		sizeType := iniGroup.Entry["Type"]
		if sizeType == "Fixed" {
			minSize = size
			maxSize = size
		} else if sizeType == "Scalable" {
			if !(minSizeGiven && maxSizeGiven) {
				fmt.Fprintln(os.Stderr, "Error in ", theme.Name, ", ", iniGroup.Name, " - with Type 'Scalable' both MinSize and MaxSize must be given")
				continue
			}
		} else if sizeType == "Threshold" {
			minSize = size - threshold
			maxSize = size + threshold
		} else {
			fmt.Fprintln(os.Stderr, "Error in ", theme.Name, ", ", iniGroup.Name, ", type must be given as 'Fixed', 'Scalable' or 'Threshold'")
			continue
		}

		theme.IconDirs = append(theme.IconDirs, IconDir{iniGroup.Name, minSize, maxSize, iniGroup.Entry["Context"]})
	}

	theme.Icons = make(map[string][]Icon)

	return theme, nil
}

func getAncestors(themeId string, visited common.StringSet, themeMap map[string]Theme) []string {
	ancestors := make([]string, 0)
	if themeId != "hicolor" && !visited[themeId] {
		visited.Add(themeId)
		if theme, ok := themeMap[themeId]; ok {
			ancestors = append(ancestors, themeId)
			for _,parentId := range theme.Inherits {
				ancestors = append(ancestors, getAncestors(parentId, visited, themeMap)...)
			}
		}
	}

	return ancestors
}


func collectIcons(icons map[string][]Icon, iconDirPath string, iconDir IconDir) {
	for _, ending := range []string{"png", "svg", "xpm"} {
		iconFilePaths, err := filepath.Glob(iconDirPath + "/*." + ending)
		if err != nil { panic(err) }

		for _, iconFilePath := range iconFilePaths {
			iconName := filepath.Base(iconFilePath[0:len(iconFilePath)-4])
			icon := Icon{iconName, iconDir.Context, iconDir.MinSize, iconDir.MaxSize, iconFilePath}
			iconList, ok := icons[iconName]
			if !ok {
				iconList = make([]Icon, 0)
			}
			icons[iconName] = append(iconList, icon)
		}
	}
}




func (tc Themes) FindIcon(themeId string, size uint32, name string) (Icon, bool){
	if theme, ok := tc.themes[themeId]; ok {
		for _,ancestorId := range theme.Ancestors {
			if icon, ok := FindIcon(tc.themes[ancestorId].Icons, size, name); ok {
				return icon,true
			}
		}
	}

	if icon, ok := FindIcon(tc.fallbackIcons, 0, name); ok {
		return icon, true
	}

	return Icon{}, false
}


// Somewhat inspired by pseudocode example in
//    https://specifications.freedesktop.org/icon-theme-spec/icon-theme-spec-latest.html
// Returns
//    Icon (zeroed if not found)
//    bool indicating if icon was found
func FindIcon(icons map[string][]Icon,  size uint32, iconname string) (Icon, bool ){
	shortestDistanceSoFar := uint32(math.MaxUint32)
	iconCandidate := Icon{}

	if iconList, ok := icons[iconname]; ok {
		for _, icon := range iconList {
			var distance uint32
			if icon.MinSize > size {
				distance = icon.MinSize - size
			} else if icon.MaxSize < size {
				distance = size - icon.MaxSize
			} else {
				distance = 0
			}

			if distance == 0 {
				return icon, true
			} else if distance < shortestDistanceSoFar {
				shortestDistanceSoFar = distance
				iconCandidate = icon
			}
		}
	}

	return iconCandidate, shortestDistanceSoFar < math.MaxUint32
}

// With xdg dirs at their default values, we search directories in this order:
// $HOME/.icon, $HOME/.local/share/icons, /usr/local/share/icons, /usr/share/icons, /usr/share/pixmap
// Ie. 'more local' takes precedence. eg:
// If both $HOME/.local/share/icons/hicolor/22x22/apps/myIcon.png and /usr/share/icons/hicolor/22x22/apps/myIcon.png
// exists, we prefer the one under $HOME/.local
func getSearchDirectories() []string {
	searchDirs := []string{xdg.Home() + "/.icons", xdg.DataHome() + "/icons"}
	for _,datadir := range common.Reverse(xdg.DataDirs()) {
		searchDirs = append(searchDirs, datadir + "/icons")
	}
	searchDirs = append(searchDirs, "/usr/share/pixmaps")
	return searchDirs
}


func readUint32(uintAsString string) (uint32, bool) {
	res, err := strconv.ParseUint(uintAsString, 10, 32)
	return uint32(res), err == nil
}

func readUint32OrFallback(uintAsString string, fallback uint32) uint32 {
	if res, ok := readUint32(uintAsString); ok {
		return res
	} else {
		return fallback
	}
}


