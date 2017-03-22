package service

import (
	"net/http"
	"fmt"
	"sync"
	"github.com/surlykke/RefudeServices/xdg"
	"net"
)

// NotifierPath is reserved, Get requests to this path will
// be answered with a server-sent-event stream. Attemts to map
// a resource to NotifierPath will panic
const NotifierPath = "/notify"

var	resources  map[string]Resource = make(map[string]Resource)
var	notifier   Notifier = MakeNotifier()
var mutex      sync.Mutex


type Resource interface {
	Data(r *http.Request) (int, string, []byte)
}


func Map(path string, res Resource) {
	mutex.Lock()
	defer mutex.Unlock()

	resources[path] = res
	notifier.Notify("resource-added", path)
}

func Remap(path string, res Resource) {
	mutex.Lock()
	defer mutex.Unlock()

	if _,ok := resources[path]; ok {
		resources[path] = res
		notifier.Notify("resource-updated", path)
	}
}

func Unmap(path string) {
	mutex.Lock()
	mutex.Unlock()

	if _,ok := resources[path]; ok {
		delete(resources, path)
		notifier.Notify("resource-removed", path)
	}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request for ", r.URL.Path)
	if r.URL.Path == "/notify" {
		notifier.ServeHTTP(w, r)
	} else {
		mutex.Lock()
		defer mutex.Unlock()

		if res,ok := resources[r.URL.Path]; ok {
			statusCode, contentType, bytes := res.Data(r)
			if statusCode != http.StatusOK {
				w.WriteHeader(statusCode)
			}
			if contentType != "" {
				w.Header().Set("Content-Type", contentType)
			}
			if bytes != nil {
				w.Write(bytes)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func Serve(socketName string) {
	socketPath := xdg.RuntimeDir() + "/" + socketName

	if listener,err := net.ListenUnix("unix", &net.UnixAddr{Name: socketPath, Net: "unix"}); err != nil {
		panic(err)
	} else {
		http.Serve(listener, http.HandlerFunc(ServeHTTP))
	}

}



