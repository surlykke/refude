package repo

import (
	"cmp"
	"fmt"
	"sync"

	"github.com/surlykke/RefudeServices/lib/entity"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/response"
)

type SyncMap[K cmp.Ordered, V entity.Servable] struct {
	m        map[K]V
	lock     sync.Mutex
	basepath string
}

func MakeSynkMap[K cmp.Ordered, V entity.Servable]() *SyncMap[K, V] {
	var m = &SyncMap[K, V]{
		m: make(map[K]V),
	}

	return m
}

// Called before trafic begins
func (this *SyncMap[K, V]) SetPrefix(basepath string) {
	this.basepath = basepath
	this.setPaths()
}

func (this *SyncMap[K, V]) Get(k K) (V, bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
	v, ok := this.m[k]
	return v, ok
}

func (this *SyncMap[K, V]) Put(k K, v V) {
	this.lock.Lock()
	defer this.lock.Unlock()
	v.GetBase().Path = fmt.Sprintf("%s%v", this.basepath, k)
	v.GetBase().BuildLinks()
	this.m[k] = v
}

func (this *SyncMap[K, V]) Remove(k K) (V, bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
	v, ok := this.m[k]
	if ok {
		delete(this.m, k)
	}
	return v, ok
}

func (this *SyncMap[K, V]) Replace(newSet map[K]V) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.m = newSet
	this.setPaths()
}

func (this *SyncMap[K, V]) GetAll() []V {
	this.lock.Lock()
	defer this.lock.Unlock()
	var list = make([]V, 0, len(this.m))
	for _, v := range this.m {
		list = append(list, v)
	}
	return list
}

func (this *SyncMap[K, V]) GetForSearch(sink *[]entity.Base) {
	for _, v := range this.GetAll() {
		if !v.OmitFromSearch() {
			*sink = append(*sink, *v.GetBase())
		}
	}
}

func (this *SyncMap[K, V]) DoGetSingle(id K) response.Response {
	if v, ok := this.Get(id); ok {
		return response.Json(v)
	} else {
		return response.NotFound()
	}
}

func (this *SyncMap[K, V]) DoGetAll() response.Response {
	fmt.Println("GetAllHandler")
	return response.Json(this.GetAll())
}

func (this *SyncMap[K, V]) DoPost(id K, action string) response.Response {
	if v, ok := this.Get(id); !ok {
		return response.NotFound()
	} else if postable, ok := any(v).(link.Postable); !ok {
		return response.NotAllowed()
	} else {
		return postable.DoPost(action)
	}
}

func (this *SyncMap[K, V]) GetPaths() []string {
	var paths = make([]string, 0, len(this.m))
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, v := range this.m {
		paths = append(paths, v.GetBase().Path)
	}
	return paths
}

func (this *SyncMap[K, V]) setPaths() {
	for k, v := range this.m {
		v.GetBase().Path = fmt.Sprintf("%s%v", this.basepath, k)
		v.GetBase().BuildLinks()
	}
}
