// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package service

import (
	"net/http"
	"log"
	"context"
	"net"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"syscall"
	"fmt"
	"github.com/surlykke/RefudeServices/lib/requestutils"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/query"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"encoding/json"
)

func seemsToBeRunning(socketPath string) bool {
	client := http.Client{
		Transport: &http.Transport{

			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	if response, err := client.Get("http://localhost/ping"); err == nil {
		response.Body.Close()
		return true
	} else {
		return false
	}
}

func makeListener(socketName string) (*net.UnixListener, bool) {
	socketPath := xdg.RuntimeDir + "/" + socketName

	if seemsToBeRunning(socketPath) {
		log.Fatal("Application seems to be running. Let's leave it at that")
	}

	syscall.Unlink(socketPath)

	listener, err := net.ListenUnix("unix", &net.UnixAddr{
		Name: socketPath,
		Net:  "unix",
	});
	if err != nil {
		fmt.Println(err)
		return nil, false
	} else {
		return listener, true
	}
}

func Serve(socketName string, jsonCollection JsonCollection) {
	if listener, ok := makeListener(socketName); ok {
		http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sp := standardize(r.URL.Path)
			if sp == "/search" {
				if r.Method == "GET" {
					Search(w, r, jsonCollection)
				} else {
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			} else if sp == "/links" {
				var links = jsonCollection.GetLinks()
				if bytes, err := json.Marshal(links); err == nil {
					w.Header().Set("Content-Type", "application/json")
					w.Write(bytes)
				} else {
					requestutils.ReportUnprocessableEntity(w, resource.ToJSon(err))
				}
			} else if res := jsonCollection.GetResource(sp); res == nil {
				w.WriteHeader(http.StatusNotFound)
			} else {
				res.ServeHTTP(w, r)
			}
		}))
	}
}

func ServeWith(socketName string, handler http.Handler) {
	if listener, ok := makeListener(socketName); ok {
		http.Serve(listener, handler)
	}
}

func getSearchParams(w http.ResponseWriter, r *http.Request) (mediatype.MediaType, query.Matcher, error) {
	var matcher query.Matcher
	var flatParams map[string]string
	var err error
	if flatParams, err = requestutils.GetSingleParams(r, "type", "q"); err != nil {
		requestutils.ReportUnprocessableEntity(w, resource.ToJSon(err))
		return "", nil, err
	}
	if q, ok := flatParams["q"]; ok {
		if matcher, err = query.Parse(q); err != nil {
			fmt.Println("Parsing problem:", err)
			requestutils.ReportUnprocessableEntity(w, resource.ToJSon(err))
			return "", matcher, err
		}
	}

	return mediatype.MediaType(flatParams["type"]), matcher, err
}

func Search(w http.ResponseWriter, r *http.Request, jsonCollection JsonCollection) {
	if mt, matcher, err := getSearchParams(w, r); err == nil {
		var allResources = jsonCollection.GetAll();
		if mt != "" {
			var tmp = make([]*resource.JsonResource, len(allResources))
			var found = 0;
			for _, jsonRes := range allResources {
				if mediatype.MediaTypeMatch(mt, jsonRes.GetMt()) {
					tmp[found] = jsonRes
					found++
				}
				allResources = tmp[:found]
			}
		}

		if matcher != nil {
			var tmp = make([]*resource.JsonResource, len(allResources))
			var found = 0;
			for _, jsonRes := range allResources {
				if jsonRes.Matches(matcher) {
					tmp[found] = jsonRes
					found++
				}
			}
			allResources = tmp[:found]
		}

		if bytes, err := json.Marshal(allResources); err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write(bytes)
		} else {
			panic(fmt.Sprintln("Problem marshalling searchresult: ", err))
		}
	} else {
		requestutils.ReportUnprocessableEntity(w, resource.ToJSon(err))
	}
}
