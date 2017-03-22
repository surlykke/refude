package main

import (
	"net/http"
	"github.com/surlykke/RefudeServices/service"
	"github.com/surlykke/RefudeServices/xdg"
	"net"
)

func main() {
	service.Setup()
	wm := WindowManager{}
	go wm.Run()

	socketPath := xdg.RuntimeDir() + "/org.refude.wm-service"

	if listener,err := net.ListenUnix("unix", &net.UnixAddr{socketPath, "unix"}); err != nil {
		panic(err)
	} else {
		http.Serve(listener, http.HandlerFunc(service.ServeHTTP))
	}

}