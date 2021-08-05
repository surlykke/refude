package search

import (
	"net/http"
	"sort"
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/statusnotifications"
	"github.com/surlykke/RefudeServices/windows"
)

type SearchResult struct {
	Links      []resource.Link `json:"_links"`
	RefudeType string          `json:"refudeType"`
	Term       string          `json:"term"`
}

func makeSearchResult(term string) SearchResult {
	var sr = SearchResult{
		Links:      make([]resource.Link, 0, 1000),
		RefudeType: "search",
		Term:       term,
	}
	sr.Links = append(sr.Links, resource.MakeLink("/desktop/search?term="+term, "Desktop Search", "", relation.Self))
	return sr
}

type rankedLink struct {
	resource.Link
	rank int
}

func makeRankedLink(href, title, iconName, refudeType string, rank int) rankedLink {
	return rankedLink{
		Link: resource.Link{
			Href:       href,
			Title:      title,
			Icon:       resource.IconUrl(iconName),
			Relation:   relation.Related,
			RefudeType: refudeType,
		},
		rank: rank,
	}
}

type rankedLinks []rankedLink

// -------- sort.Interface ----------------------

func (se rankedLinks) Len() int { return len(se) }

func (se rankedLinks) Less(i int, j int) bool {
	if se[i].rank == se[j].rank {
		return se[i].Href < se[j].Href // Not that Href is particularly relevant, sortingwise. Just to have a reproducible order
	} else {
		return se[i].rank < se[j].rank
	}
}

func (se rankedLinks) Swap(i int, j int) { se[i], se[j] = se[j], se[i] }

// ----------- sort.Interface end ---------------

func addRankedLinks(sr *SearchResult, links []rankedLink) {
	for _, rl := range links {
		sr.Links = append(sr.Links, rl.Link)
	}
}

func collectSearchResult(term string) SearchResult {
	var termRunes = []rune(term)
	var sr = makeSearchResult(term)

	var temp = make(rankedLinks, 0, 1000)

	var doCrawl = func(crawl searchutils.Crawl, refudeType string, doSort bool) {
		var crawler = func(href, title, iconName string, keywords ...string) {
			var titleRunes = []rune(strings.ToLower(title))
			if rnk := searchutils.FluffyIndex(titleRunes, termRunes); rnk > -1 {
				temp = append(temp, makeRankedLink(href, title, iconName, refudeType, rnk))
			} else if len(termRunes) >= 3 {
				for _, kw := range keywords {
					if strings.Index(kw, term) > -1 {
						temp = append(temp, makeRankedLink(href, title, iconName, refudeType, 1000))
						break
					}
				}
			}
		}

		crawl(term, true, crawler)
		if doSort {
			sort.Sort(temp)
		}
		addRankedLinks(&sr, temp)
		temp = temp[:0]
	}

	doCrawl(notifications.Crawl, "notification", false)
	doCrawl(windows.Crawl, "window", false)

	if len(term) > 0 {
		doCrawl(applications.Crawl, "application", true)
		if len(term) > 3 {
			doCrawl(file.Crawl, "file", true)
			doCrawl(power.Crawl, "device", true)
		}
	}

	return sr
}

func collectPaths(prefix string) []string {
	var paths []string = make([]string, 0, 1000)
	var crawler = func(href, title, iconName string, keywords ...string) {
		if strings.HasPrefix(href, prefix) {
			paths = append(paths, href)
		}
	}
	windows.Crawl("", false, crawler)
	applications.Crawl("", false, crawler)
	icons.Crawl("", false, crawler)
	statusnotifications.Crawl("", false, crawler)
	notifications.Crawl("", false, crawler)
	power.Crawl("", false, crawler)

	for _, path := range []string{"/icon", "/search/desktop", "/search/paths", "/watch", "/doc"} {
		if strings.HasPrefix(path, prefix) {
			paths = append(paths, path)
		}
	}
	return paths
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/search/desktop" {
		if r.Method == "GET" {
			respond.AsJson(w, collectSearchResult(requests.Term(r)))
		} else {
			respond.NotAllowed(w)
		}
	} else if r.URL.Path == "/search/paths" {
		if r.Method == "GET" {
			respond.AsJson(w, collectPaths(requests.GetSingleQueryParameter(r, "prefix", "")))
		} else {
			respond.NotAllowed(w)
		}
	} else {
		respond.NotFound(w)
	}
}
