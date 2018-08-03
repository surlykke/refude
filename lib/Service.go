// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package lib

import (
	"net/http"
	"log"
	"context"
	"net"
	"syscall"
	"fmt"
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
	socketPath := RuntimeDir + "/" + socketName

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
			sp := Standardize(r.URL.Path)
			if sp == "/search" {
				if r.Method == "GET" {
					Search(w, r, jsonCollection)
				} else {
					w.WriteHeader(http.StatusMethodNotAllowed)
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

func getSearchParams(w http.ResponseWriter, r *http.Request) (MediaType, Matcher, error) {
	var matcher Matcher
	var flatParams map[string]string
	var err error
	if flatParams, err = GetSingleParams(r, "type", "q"); err != nil {
		return "", nil, err
	} else if q, ok := flatParams["q"]; ok {
		if matcher, err = Parse(q); err != nil {
			fmt.Println("Parsing problem:", err)
			return "", matcher, err
		}
	}

	return MediaType(flatParams["type"]), matcher, err
}

func Search(w http.ResponseWriter, r *http.Request, jsonCollection JsonCollection) {
	fmt.Println("Search, query:", r.URL.RawQuery);
	if mt, matcher, err := getSearchParams(w, r); err == nil {
		var allResources = jsonCollection.GetAll();
		if mt != "" {
			var tmp = make([]*JsonResource, len(allResources))
			var found = 0;
			for _, jsonRes := range allResources {
				if MediaTypeMatch(mt, jsonRes.GetMt()) {
					tmp[found] = jsonRes
					found++
				}
				allResources = tmp[:found]
			}
		}

		if matcher != nil {
			var tmp = make([]*JsonResource, len(allResources))
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
		ReportUnprocessableEntity(w, err)
	}
}
