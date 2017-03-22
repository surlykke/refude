package main

import (
	"github.com/surlykke/RefudeServices/service"
)

func main() {
	var iconService IconService
	iconService.update()
	service.Serve("org.refude.icon-service")
}
