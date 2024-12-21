package browser

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type Bookmark struct {
	resource.ResourceData
	Id          string
	ExternalUrl string
}

func (this *Bookmark) DoPost(w http.ResponseWriter, r *http.Request) {
	xdg.RunCmd("xdg-open", this.ExternalUrl)
}

// We use this for icon url
// https://t0.gstatic.com/faviconV2?client=SOCIAL&type=FAVICON&url=<bookmark url>
