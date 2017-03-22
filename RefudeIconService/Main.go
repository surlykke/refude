package main

import (
	"net/http"
	"github.com/surlykke/RefudeServices/xdg"
	"net"
)

func main() {
	var iconService IconService
	iconService.update()

	socketPath := xdg.RuntimeDir() + "/org.refude.icon-service"

	if listener,err := net.ListenUnix("unix", &net.UnixAddr{socketPath, "unix"}); err != nil {
		panic(err)
	} else {
		http.Serve(listener, iconService)
	}
}
