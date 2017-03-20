package main

import (
	"github.com/surlykke/RefudeServices/service"
	"net/http"
)

func main() {
	service.Setup()
	pm := 	&PowerManager{}
	go pm.Run()
	http.ListenAndServe(":8000", http.HandlerFunc(service.ServeHTTP))
}
