package icons

import (
	"net/http"
	"strings"
)

func Handler(r *http.Request) http.Handler {
	lock.Lock()
	defer lock.Unlock()
	if strings.HasPrefix(r.URL.Path, "/icon/") {
		var iconName = r.URL.Path[6:]
		for _, theme := range themeList {
			if icon, ok := theme.icons[iconName]; ok {
				return icon
			}
		}
		if icon, ok := hicolor.icons[iconName]; ok {
			return icon
		}
		if image, ok := other[iconName]; ok {
			return image
		}
	}
	return nil
}

func AllPaths() []string {
	// TODO
	/*lock.Lock()
	defer lock.Unlock()
	var paths = make([]string, 0, len(themes)+1)
	for path := range themes {
		paths = append(paths, path)
	}
	paths = append(paths, "/icon")
	return paths*/
	return []string{""}
}
