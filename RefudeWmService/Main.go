package main

import (
	"github.com/surlykke/RefudeServices/service"
)

func main() {
	wm := WindowManager{}
	go wm.Run()
	service.Serve("org.refude.wm-service")
}