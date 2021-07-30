package search

import (
	"net/http"
	"sort"
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/statusnotifications"
	"github.com/surlykke/RefudeServices/windows"
)

type SearchResult struct {
	respond.Resource
	Term string `json:"term"`
}

func makeSearchResult(term string) SearchResult {
	var sr = SearchResult{}
	sr.Links = make([]respond.Link, 0, 1000)
	sr.AddSelfLink("/desktop/search?term="+term, "Desktop Search", "")
	sr.Traits = []string{"search"}
	sr.Term = term
	return sr
}

type rankedLink struct {
	respond.Link
	rank int
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
	var crawler = func(res *respond.Resource, keywords []string) {
		var link = res.GetRelatedLink()
		var titleRunes = []rune(strings.ToLower(link.Title))
		if rnk := searchutils.FluffyIndex(titleRunes, termRunes); rnk > -1 {
			temp = append(temp, rankedLink{link, rnk})
		} else if len(termRunes) >= 3 {
			for _, kw := range keywords {
				if strings.Index(kw, term) > -1 {
					temp = append(temp, rankedLink{link, 1000})
					break
				}
			}
		}
	}

	var doCrawl = func(crawl searchutils.Crawl, doSort bool) {
		crawl(term, true, crawler)
		if doSort {
			sort.Sort(temp)
		}
		addRankedLinks(&sr, temp)
		temp = temp[:0]
	}

	doCrawl(notifications.Crawl, false)
	doCrawl(windows.Crawl, false)

	if len(term) > 0 {
		doCrawl(applications.Crawl, true)
		if len(term) > 3 {
			doCrawl(file.Crawl, true)
			doCrawl(power.Crawl, true)
		}
	}

	return sr
}

func collectPaths(prefix string) []string {
	var paths []string = make([]string, 0, 1000)
	var crawler = func(res *respond.Resource, keywords []string) {
		if strings.HasPrefix(res.Self().Href, prefix) {
			paths = append(paths, res.Self().Href)
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
