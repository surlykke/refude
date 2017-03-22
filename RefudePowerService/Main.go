package main

import (
	"github.com/surlykke/RefudeServices/service"
)

func main() {
	pm := 	&PowerManager{}
	go pm.Run()
	service.Serve("org.refude.power-service")
}
