package main

import "github.com/surlykke/RefudeServices/lib/service"

func main() {
	go consumeSignals()
	go StatusNotifierWatcher(registerChannel, unregisterChannel)
	service.Serve("org.refude.statusnotifier-service")
}
