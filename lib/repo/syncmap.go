package repo

import (
	"cmp"
	"fmt"
	"sync"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/response"
)

type Storable interface {
	GetPath() string
	SetPath(string)
}

type SyncMap[K cmp.Ordered, V Storable] struct {
	m        map[K]V
	lock     sync.Mutex
	basepath string
}

func MakeSynkMap[K cmp.Ordered, V Storable]() *SyncMap[K, V] {
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
	v.SetPath(fmt.Sprintf("%s%v", this.basepath, k))
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

func (this *SyncMap[K, V]) DoGetSingle(id K) response.Response {
	fmt.Print("SyncMap.DoGetSingle, id:'", id, "'\n")
	fmt.Println("GetHandler, id:", id)
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
	fmt.Print("SyncMap.DoPost, id:", id, "'\n")
	if v, ok := this.Get(id); !ok {
		fmt.Println("Not found")
		return response.NotFound()
	} else if postable, ok := any(v).(link.Postable); !ok {
		fmt.Println("Not postable")
		return response.NotAllowed()
	} else {
		fmt.Println("Doing it...")
		return postable.DoPost(action)
	}
}

func (this *SyncMap[K, V]) GetPaths() []string {
	var paths = make([]string, 0, len(this.m))
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, v := range this.m {
		paths = append(paths, v.GetPath())
	}
	return paths
}

func (this *SyncMap[K, V]) setPaths() {
	for k, v := range this.m {
		v.SetPath(fmt.Sprintf("%s%v", this.basepath, k))
	}
}
