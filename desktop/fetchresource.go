package desktop

import (
	"strings"

	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/resourcerepo"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/start"
)

type LinkCollection struct {
	Profiles []string
	links map[string]link.List
}

func (lc LinkCollection) Links(profile string) link.List {
	return lc.links[profile]
}

func fetchResourceData(resourcePath string, term string) (map[string]any, bool) {
	if res := fetchResource(resourcePath); res != nil {
		return map[string]any{
			"Title": res.GetTitle(),
			"Icon": res.GetIconUrl(),
			"Actions": res.Actions(),
			"LinkCollection": buildLinkCollection(res, term),
			"Term": term,
		}, true
		
	} else {
		return nil, false
	}
}

// TODO Move to resource.Resource
func buildLinkCollection(res resource.Resource, searchTerm string) LinkCollection {
	var collection = LinkCollection{Profiles: make([]string, 0, 8), links: make(map[string]link.List, 6)}
	var add = func(profile string, l link.Link) {
		if _, ok := collection.links[profile]; !ok {
			collection.Profiles = append(collection.Profiles, profile)
			collection.links[profile] = make(link.List, 0, 10)
		}
		collection.links[profile] = append(collection.links[profile], l)	
	}

	for _, action := range res.Actions() {
		var href = res.GetPath()
		if action.Name != "" {
			href += "?action=" + action.Name
		}
		if searchutils.Match(searchTerm, action.Name) < 0 {
			continue
		}
		add("action", link.Make(href, action.Title, action.IconUrl, relation.Action))
	}
	if deleteTitle, ok := res.DeleteAction(); ok {
		if searchutils.Match(searchTerm, deleteTitle) > -1 {
			add("action", link.Make(res.GetPath(), deleteTitle, "", relation.Delete))
		}
	}

	for _, lnk := range res.Links(searchTerm) {
		add(lnk.Profile, lnk)
	}

	return collection
}

func fetchResource(path string) resource.Resource {
	if path == "/start" {
		return start.Start
	} else if strings.HasPrefix(path, "/file/") {
		return file.FileRepo.GetResource(path[5:])
	} else if res, ok := resourcerepo.Get(path); ok {
		return res 
	} else {
		return nil
	}
}

func fetchResourceFromCollection[T resource.Resource](collection *resource.Collection[T], path string) resource.Resource {
	if t, ok := collection.Get(path); ok {
		return t
	} else {
		return nil
	}
}

