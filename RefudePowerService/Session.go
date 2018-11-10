package main

import (
	"github.com/surlykke/RefudeServices/lib/resource"
)

const SessionMediaType resource.MediaType = "application/vnd.org.refude.session+json"

type Session struct {
	resource.AbstractResource
}