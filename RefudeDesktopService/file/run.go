package file

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/surlykke/RefudeServices/lib/xdg"

	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
)

func init() {
	if err := magicmime.Open(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR); err != nil {
		panic(err)
	}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var path = requests.GetSingleQueryParameter(r, "path", "")
	if path == "" {
		respond.UnprocessableEntity(w, fmt.Errorf("A nonempty path must be given"))
		return
	}

	if !strings.HasPrefix(path, "/") {
		path = xdg.Home + "/" + path
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		respond.NotFound(w)
	} else if err != nil {
		respond.ServerError(w, err)
	} else {
		var mimetype, _ = magicmime.TypeByFile(path)
		var defaultApp = "TODO"
		respond.AsJson(w, r, MakeFile(path, mimetype, defaultApp).ToStandardFormat())
	}

}
