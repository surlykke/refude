package main

import (
	"github.com/surlykke/RefudeServices/service"
	"net/http"
	"github.com/surlykke/RefudeServices/xdg"
	"net"
)

func main() {
	pm := 	&PowerManager{}
	go pm.Run()
	service.Serve("org.refude.power-service")
}
