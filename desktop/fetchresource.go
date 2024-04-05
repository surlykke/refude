package desktop

import (
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/search"
)

var headingOrder = []string{"Actions", "Notifications", "Windows", "Tabs", "Applications", "Files", "Other"}

var profileHeadingMap = map[string]string{
	"notification": "Notifications", "window": "Windows", "browsertab": "Tabs", "application": "Applications", "file": "Files",
}

func collectTemplateData(resourcePath string, term string) (map[string]any, bool) {
	if res := search.FetchResource(resourcePath); res != nil {
		var m = map[string]any{
			"Searchable": res.Base().Searchable(),
			"Title":      res.Base().Title,
			"Icon":       res.Base().IconUrl,
			"Term":       term,
			"Actions":    res.Base().ActionLinks(),
		}

		var resources = make(map[string][]resource.Resource, 9)
		for _, subRes := range res.Search(term) {
			var heading string
			var ok bool
			if heading, ok = profileHeadingMap[subRes.Base().Profile]; !ok {
				heading = "Other"
			}
			resources[heading] = append(resources[heading], subRes)
		}
		m["Resources"] = resources

		var headings = make([]string, 0, 9)
		for _, heading := range headingOrder {
			if _, ok := resources[heading]; ok {
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
	} else if heading, ok := profileHeadingMap[l.Profile]; ok {
		return heading
	} else {
		return "Other"
	}
}
