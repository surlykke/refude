package main

import (
	"github.com/surlykke/RefudeServices/service"
	"net/http"
	"github.com/surlykke/RefudeServices/xdg"
	"net"
)

func main() {
	service.Setup()
	pm := 	&PowerManager{}
	go pm.Run()

	socketPath := xdg.RuntimeDir() + "/org.refude.power-service"

	if listener,err := net.ListenUnix("unix", &net.UnixAddr{socketPath, "unix"}); err != nil {
		panic(err)
	} else {
		http.Serve(listener, http.HandlerFunc(service.ServeHTTP))
	}
}
