package main

import "github.com/surlykke/RefudeServices/lib/service"

func main() {
	go runWatcher()
	service.Serve("org.refude.statusnotifier-service")
}
