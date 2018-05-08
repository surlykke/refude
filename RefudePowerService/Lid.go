package main

import (
	"github.com/surlykke/RefudeServices/lib/mediatype"
)

const LidMediaType mediatype.MediaType = "application/vnd.org.refude.upowerlid+json"

type Lid struct {
	Open bool
}

