// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package lib

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"syscall"

	"github.com/surlykke/RefudeServices/lib/xdg"
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
		_ = response.Body.Close()
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

	_ = syscall.Unlink(socketPath)

	listener, err := net.ListenUnix("unix", &net.UnixAddr{
		Name: socketPath,
		Net:  "unix",
	})
	if err != nil {
		fmt.Println(err)
		return nil, false
	} else {
		return listener, true
	}
}

func Serve(socketName string, handler http.Handler) {
	if seemsToBeRunning(socketName) {
		fmt.Println("Something is already running on", socketName)
	} else if listener, ok := makeListener(socketName); ok {
		http.Serve(listener, handler)
	} else {
		fmt.Println("Unable to listen on", socketName)
	}
}
