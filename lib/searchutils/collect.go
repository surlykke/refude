package searchutils

import (
	"sort"
	"strings"

	"github.com/surlykke/RefudeServices/lib/respond"
)

type RankedResource struct {
	*respond.StandardFormat
	Rank int
}

type RankedList []RankedResource

func (rL RankedList) Len() int               { return len(rL) }
func (rL RankedList) Less(i int, j int) bool { return rL[i].Rank < rL[j].Rank }
func (rL RankedList) Swap(i int, j int)      { rL[i], rL[j] = rL[j], rL[i] }

type Collector struct {
	term            string
	termAsRunes     []rune
	rankedResources RankedList
	omitNoDisplay   bool
}

func MakeCollector(term string, initialCapacity int, omitNoDisplay bool) *Collector {
	var lowercaseTerm = strings.ToLower(term)
	return &Collector{
		term:            lowercaseTerm,
		termAsRunes:     []rune(lowercaseTerm),
		rankedResources: make(RankedList, 0, initialCapacity),
		omitNoDisplay:   omitNoDisplay,
	}
}

func (c *Collector) Collect(resource *respond.StandardFormat) {
	if c.omitNoDisplay && resource.NoDisplay {
		return
	}

	var rank = Rank(resource, c.term, c.termAsRunes)
	if rank > -1 {
		c.rankedResources = append(c.rankedResources, RankedResource{resource, rank})
	}
}

func (c *Collector) SortByRankAndGet() []*respond.StandardFormat {
	sort.Sort(c.rankedResources)
	return c.Get()
}

func (c *Collector) Get() []*respond.StandardFormat {
	var resources = make([]*respond.StandardFormat, 0, len(c.rankedResources))
	for _, rankedResource := range c.rankedResources {
		resources = append(resources, rankedResource.StandardFormat)
	}

	return resources
}

func (c *Collector) Clear() {
	c.rankedResources = c.rankedResources[0:0]
}
