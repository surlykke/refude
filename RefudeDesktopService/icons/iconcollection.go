package icons

import (
	"fmt"
	"log"
	"math"
	"path/filepath"
	"sort"
	"sync"

	"github.com/surlykke/RefudeServices/lib/resource"

	"github.com/surlykke/RefudeServices/lib/slice"
)

type (
	themeIconImage struct {
		iconName string
		themeId  string
		iconDir  string
		path     string
	}

	sessionIconImage struct {
		name string
		size uint32
	}

	otherIconImage struct {
		name string
		path string
	}
)

var (
	lock         = &sync.Mutex{}
	themes       = make(map[string]*IconTheme)
	themeIcons   = make(map[string]map[string][]IconImage) // themeid -> iconname -> []IconImage
	sessionIcons = make(map[string][]IconImage)
	otherIcons   = make(map[string]IconImage) // name -> IconImage

	themeIconImages   = make(chan themeIconImage)
	sessionIconImages = make(chan sessionIconImage)
	otherIconImages   = make(chan otherIconImage)
)

func initIconCollection(baseDirs []string) {
	for _, baseDir := range baseDirs {
		if indexFilePaths, err := filepath.Glob(baseDir + "/*/index.theme"); err != nil {
			log.Println(err)
			continue
		} else {
			for _, indexFilePath := range indexFilePaths {
				if theme, ok := readTheme(indexFilePath); ok {
					if _, ok = themes[theme.Id]; !ok {
						theme.Links = resource.Links{Self: "/icontheme/" + theme.Id, RefudeType: "icontheme"}
						themes[theme.Id] = theme
						themeIcons[theme.Id] = make(map[string][]IconImage)
					}
				}
			}
		}
	}

	// For each theme, calculate search order - ie. order in which to search parent themes
	for themeId, theme := range themes {
		if themeId != "hicolor" {

			theme.SearchOrder = []string{themeId}

			for i := 0; i < len(theme.SearchOrder); i++ {
				for _, inheritId := range themes[theme.SearchOrder[i]].Inherits {
					if inheritId == "hicolor" {
						continue
					} else if slice.Contains(theme.SearchOrder, inheritId) {
						continue
					} else {
						theme.SearchOrder = append(theme.SearchOrder, inheritId)
					}
				}
			}
		}
		theme.SearchOrder = append(theme.SearchOrder, "hicolor")
	}

	var themeResources = make(map[string]resource.Resource)
	var themeResourceList resource.ResourceList
	var themePaths resource.PathList
	for themeId, theme := range themes {
		themeResources["/icontheme/"+themeId] = theme
		themeResourceList = append(themeResourceList, theme)
		themePaths = append(themePaths, theme.GetSelf())
	}
	sort.Sort(themeResourceList)
	sort.Sort(themePaths)
	themeResources["/iconthemes"] = themeResourceList
	themeResources["/iconthemepaths"] = themePaths
	resource.MapCollection(&themeResources, "iconthemes")

	resource.MapSingle("/icon", &IconResource{})
}

func recieveImages() {
	for {
		select {
		case image := <-themeIconImages:
			if theme, ok := themes[image.themeId]; !ok {
				// Ignore
			} else if iconDir, ok := theme.Dirs[image.iconDir]; !ok {
				// Ignore
			} else {
				var id, name = image.themeId, image.iconName
				themeIcons[id][name] = append(themeIcons[id][name], IconImage{
					Context: iconDir.Context,
					MinSize: iconDir.MinSize,
					MaxSize: iconDir.MaxSize,
					Path:    image.path,
				})
			}
		case image := <-sessionIconImages:
			var path = fmt.Sprintf("%s/%d/%s.png", refudeSessionIconsDir, image.size, image.name) // session icons always png's
			sessionIcons[image.name] = append(sessionIcons[image.name], IconImage{
				MinSize: image.size,
				MaxSize: image.size,
				Path:    path,
			})
		case image := <-otherIconImages:
			otherIcons[image.name] = IconImage{Path: image.path}
		}
	}
}

func findImage(themeId string, iconName string, size uint32) (IconImage, bool) {
	lock.Lock()
	defer lock.Unlock()
	if theme, ok := themes[themeId]; !ok {
		return IconImage{}, false
	} else {
		for _, id := range theme.SearchOrder {
			if imageList, ok := themeIcons[id][iconName]; ok {
				return findBestMatch(imageList, size), true
			}
		}

		if imageList, ok := sessionIcons[iconName]; ok {
			return findBestMatch(imageList, size), true
		}

		image, ok := otherIcons[iconName]
		return image, ok
	}
}

func findBestMatch(images []IconImage, size uint32) IconImage {
	var shortestDistanceSoFar = uint32(math.MaxUint32)
	var candidate IconImage

	for _, img := range images {
		var distance uint32
		if img.MinSize > size {
			distance = img.MinSize - size
		} else if img.MaxSize < size {
			distance = size - img.MaxSize
		} else {
			distance = 0
		}

		if distance < shortestDistanceSoFar {
			shortestDistanceSoFar = distance
			candidate = img
		}
		if distance == 0 {
			break
		}

	}

	return candidate
}
