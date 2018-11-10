// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"context"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

var prefixes = []string{
	"/desktop-service/",
	"/icon-service/",
	"/notifications-service/",
	"/power-service/",
	"/statusnotifier-service/",
	"/wm-service/",
}

func dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	socketAddr := xdg.RuntimeDir + "/org.refude." + addr[0:len(addr)-3] // Strip trailing ':80'
	return net.Dial("unix", socketAddr)
}

func director(req *http.Request) {
	for _, prefix := range prefixes {
		// We will find one that matches
		if strings.HasPrefix(req.URL.Path, prefix) {
			req.URL.Scheme = "http"
			req.URL.Host = prefix[1 : len(prefix)-1]
			req.URL.Path = req.URL.Path[len(prefix)-1:]
			return
		}
	}
}

var reverseProxy = httputil.ReverseProxy{
	Director:      director,
	Transport:     &http.Transport{DialContext: dialContext},
	FlushInterval: time.Second,
}

func handler(w http.ResponseWriter, r *http.Request) {
	for _, prefix := range prefixes {
		if strings.HasPrefix(r.URL.Path, prefix) {
			reverseProxy.ServeHTTP(w, r)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func main() {
	http.ListenAndServe(":7938", http.HandlerFunc(handler))
}
