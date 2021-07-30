package icons

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	for _, name := range dashSplit(iconName) {
		for _, theme := range themeList {
			if icon, ok := theme.icons[name]; ok {
				return icon
			}
		}
		if icon, ok := hicolor.icons[name]; ok {
			return icon
		}
	}
	return nil
}

/**
 * By the icon naming specification, dash ('-') seperates 'levels of specificity'. So given an icon name
 * 'input-mouse-usb', the levels of spcicificy, and the names and order we search will be: 'input-mouse-usb',
 * 'input-mouse' and 'input'
 */
func dashSplit(name string) []string {
	var res = make([]string, 0, 3)
	for {
		res = append(res, name)
		if pos := strings.LastIndex(name, "-"); pos > 0 {
			name = name[0:pos]
		} else {
			break
		}
	}
	return res
}

func Crawl(term string, forDisplay bool, crawler searchutils.Crawler) {
	lock.Lock()
	defer lock.Unlock()
	for _, theme := range themeMap {
		crawler(&theme.Resource, nil)
	}
}
