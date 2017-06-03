package main

import "github.com/surlykke/RefudeServices/lib/service"

func main() {
	go run()
	service.Serve("org.refude.statusnotifier-service")
}
