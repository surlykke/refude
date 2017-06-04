package main

import (
	"github.com/surlykke/RefudeServices/lib/service"
	"fmt"
)

func main() {
	// Initially, put an empty list of items up, signalling we don't have any yet
	fmt.Println("Createing empty /items/")
	service.CreateEmptyDir("/items")
	go run()
	service.Serve("org.refude.statusnotifier-service")
}
