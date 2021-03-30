package icons

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
)

func Handler(r *http.Request) http.Handler {
	lock.Lock()
	defer lock.Unlock()
	if r.URL.Path == "/icon" {
		return IconServer{}
	} else {
		return nil
	}
}

type IconServer struct{}

func (is IconServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respond.NotAllowed(w)
	} else if name := requests.GetSingleQueryParameter(r, "name", ""); name == "" {
		respond.UnprocessableEntity(w, fmt.Errorf("Query parameter 'name' must be given, and not empty"))
	} else if strings.HasPrefix(name, "/") {
		http.ServeFile(w, r, name)
	} else if icon := findIcon(name); icon != nil {
		icon.ServeHTTP(w, r)
	} else if iconImage, ok := other[name]; ok {
		iconImage.ServeHTTP(w, r)
	} else {
		respond.NotFound(w)
	}
}

func findIcon(iconName string) *Icon {

	for _, theme := range themeList {
		if icon, ok := theme.icons[iconName]; ok {
			return icon
		}
	}
	if icon, ok := hicolor.icons[iconName]; ok {
		return icon
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
