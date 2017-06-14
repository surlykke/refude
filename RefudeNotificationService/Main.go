package main

import (
	"github.com/surlykke/RefudeServices/lib/service"
	"fmt"
)

func main() {
	// Initially, put an empty list of notifications up, signalling we don't have any yet
	fmt.Println("Createing empty /notifications/")
	service.CreateEmptyDir("/notifications")
	Setup()
	service.Serve("org.refude.notifications-service")
}
