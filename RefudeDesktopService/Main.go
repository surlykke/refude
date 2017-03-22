package main

import (
	"github.com/surlykke/RefudeServices/service"
)

func main() {
	desktop := NewDesktop()
	go desktop.Run()
	service.Serve("org.refude.desktop-service")
}
