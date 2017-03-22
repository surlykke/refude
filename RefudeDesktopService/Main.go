package main

import (
	"net/http"
	"github.com/surlykke/RefudeServices/service"
	"net"
	"github.com/surlykke/RefudeServices/xdg"
)

func main() {
	service.Setup()
	desktop := NewDesktop()
	go desktop.Run()

	socketPath := xdg.RuntimeDir() + "/org.refude.desktop-service"

	if listener,err := net.ListenUnix("unix", &net.UnixAddr{socketPath, "unix"}); err != nil {
		panic(err)
	} else {
		http.Serve(listener, http.HandlerFunc(service.ServeHTTP))
	}

}
