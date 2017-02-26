package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type JsonResourceCollection struct {
	jsons map[string][]byte
	mutex sync.RWMutex
}

func (j JsonResourceCollection) GET(w http.ResponseWriter, r *http.Request) {
	j.mutex.RLock()
	jsonsCopy := j.jsons
	j.mutex.RUnlock()
	if json, ok := jsonsCopy[r.URL.Path]; ok {
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

type DesktopService struct {
	mimeMap MimeMap
	appMap AppMap
	JsonResourceCollection

}

func (d DesktopService) POST(w http.ResponseWriter, r *http.Request) {
	// FIXME
	w.WriteHeader(http.StatusAccepted)
}

func (d DesktopService) PATCH(w http.ResponseWriter, r *http.Request) {
	// FIXME
	w.WriteHeader(http.StatusAccepted)
}

func (d DesktopService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		d.GET(w, r)
	case "POST":
		d.POST(w, r)
	case "PATCH":
		d.PATCH(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (d *DesktopService) update() {
	mimeMap, appMap := CollectFromDesktop()
	mimetypePaths := make([]string, 0, len(mimeMap))
	applicationPaths := make([]string, 0, len(appMap))
	preJson := make(map[string]interface{})
	jsons := make(map[string][]byte)

	for mimetypeId, mimetype := range mimeMap {
		preJson["/mimetype/" + mimetypeId] = mimetype
		mimetypePaths = append(mimetypePaths, "mimetype/" + mimetypeId)
	}

	for desktopId, app := range appMap {
		preJson["/application/" + desktopId] = app
		applicationPaths = append(applicationPaths, "application/" + desktopId)
	}
	preJson["/applications"] = applicationPaths


	preJson["/mimetypes"] = mimetypePaths

	for path,res := range preJson {
		b, err := json.Marshal(res)
		if err == nil {
			jsons[path] = b
		} else {
			fmt.Println("Marshal error: ", err)
		}
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.mimeMap = mimeMap
	d.appMap = appMap
	d.jsons = jsons

}

func (d *DesktopService) Start() {
	d.update()
	fmt.Println(time.Now(), "Listening on port 8000")
	http.ListenAndServe(":8000", d)
}
