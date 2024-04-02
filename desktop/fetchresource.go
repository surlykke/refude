package desktop

import (
	"strings"

	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/resourcerepo"
)

var headingOrder = []string{"Actions", "Notifications", "Windows", "Tabs", "Applications", "Files", "Other"}

var profileHeadingMap = map[string]string{
	"notification": "Notifications", "window": "Windows", "browsertab": "Tabs", "application": "Applications", "file": "Files",
}

func fetchResourceData(resourcePath string, term string) (map[string]any, bool) {
	if res := fetchResource(resourcePath); res != nil {
		var m = map[string]any{
			"Title": res.GetTitle(),
			"Icon":  res.GetIconUrl(),
			"Term":  term,
		}

		var links = make(map[string]link.List, 9)
		for _, lnk := range res.Links(term) {
			links[heading(lnk)] = append(links[heading(lnk)], lnk)
		}
		m["Links"] = links

		var headings = make([]string, 0, 9)
		for _, heading := range headingOrder {
			if _, ok := links[heading]; ok {
				headings = append(headings, heading)
			}
		}
		m["Headings"] = headings

		return m, true

	} else {
		return nil, false
	}
}

func heading(l link.Link) string {
	if l.Relation == relation.Action || l.Relation == relation.Delete {
		return "Actions"
	} else if heading, ok := profileHeadingMap[l.Profile]; ok{
		return heading 
	} else { 
		return "Other"
	}
}

func fetchResource(path string) resource.Resource {
	if strings.HasPrefix(path, "/file/") {
		return file.FileRepo.GetResource(path[5:])
	} else if res, ok := resourcerepo.Get(path); ok {
		return res
	} else {
		return nil
	}
}
