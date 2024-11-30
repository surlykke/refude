package href

import (
	"strings"

	"github.com/surlykke/RefudeServices/lib/path"
)

type Href string

func Of(path path.Path) Href {
	return Href(path)
}

func (href Href) P(key string, value string) Href {
	var separator = "?"
	if strings.Index(string(href), "?") > -1 {
		separator = "&"
	}
	return href + Href(separator+key+"="+value)
}
