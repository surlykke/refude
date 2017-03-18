package main

import (
	"net/http"
	"github.com/surlykke/RefudeServices/service"
)

func main() {
	service.Setup()
	desktop := NewDesktop()
	go desktop.Run()

	http.ListenAndServe(":8000", http.HandlerFunc(service.ServeHTTP))
}
