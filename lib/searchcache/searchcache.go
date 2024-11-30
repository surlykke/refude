package searchcache

import (
	"sync"
	"time"

	"github.com/surlykke/RefudeServices/lib/resource"
)

type provider func() []resource.Link

type SearchCache struct {
	provider provider
	cached   []resource.Link
	expires  time.Time
	timout   time.Duration
	lock     sync.Mutex
}

func Make(provider provider, timeout time.Duration) *SearchCache {
	return &SearchCache{
		provider: provider,
		timout:   timeout,
	}
}

func (this *SearchCache) Get() []resource.Link {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.cached == nil || !time.Now().Before(this.expires) {
		this.cached = this.provider()
		this.expires = time.Now().Add(this.timout)
	}
	return this.cached

}
