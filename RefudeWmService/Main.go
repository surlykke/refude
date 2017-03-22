package main

import (
	"net/http"
	"github.com/surlykke/RefudeServices/service"
)

func main() {
	service.Setup()
	wm := WindowManager{}
	go wm.Run()
	http.ListenAndServe(":8000", http.HandlerFunc(service.ServeHTTP))
}