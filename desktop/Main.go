package main

import (
	"github.com/surlykke/RefudeServices/resources"
	"net/http"
)

func main() {
	resourceCollection := resources.NewResourceCollection()
	resources := CollectFromDesktop()
	resourceCollection.Set(resources)

	http.ListenAndServe(":8000", &resourceCollection)
}
