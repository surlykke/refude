package action

import (
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"net/http"
)

const ActionMediaType mediatype.MediaType = "application/vnd.org.refude.action+json"

type Executer func()

type Action struct {
	Name          string
	Comment       string
	IconName      string
	Self          string
	executer      Executer
}

func MakeAction(Name string, Comment string, IconName string, Self string, executer Executer) *Action {
	return &Action{Name, Comment, IconName, Self, executer}
}

func (a *Action) POST(w http.ResponseWriter, r *http.Request) {
	if a.executer != nil {
		a.executer()
	}
	w.WriteHeader(http.StatusAccepted)
}
