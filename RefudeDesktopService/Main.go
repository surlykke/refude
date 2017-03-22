package main

import (
	"net/http"
	"github.com/surlykke/RefudeServices/service"
	"net"
)

func main() {
	service.Setup()
	desktop := NewDesktop()
	go desktop.Run()

	if listener,err := net.ListenUnix("unix", &net.UnixAddr{"/tmp/test", "unix"}); err != nil {
		panic(err)
	} else {
		http.Serve(listener, http.HandlerFunc(service.ServeHTTP))
	}

}
