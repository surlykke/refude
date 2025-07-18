package browser

import (
	"github.com/surlykke/RefudeServices/lib/entity"
	"github.com/surlykke/RefudeServices/lib/response"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type Bookmark struct {
	entity.Base
	Id          string
	ExternalUrl string
}

func (this *Bookmark) DoPost(action string) response.Response {
	xdg.RunCmd("xdg-open", this.ExternalUrl)
	return response.Accepted()
}

// We use this for icon url
// https://t0.gstatic.com/faviconV2?client=SOCIAL&type=FAVICON&url=<bookmark url>
