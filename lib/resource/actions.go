package resource

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/requests"
)

type Executer func()

type ResourceAction struct {
	Description string
	IconName    string
	Executer    Executer `json:"-"`
}

// For embedding
type Actions struct {
	Actions map[string]ResourceAction `json:"_actions,omitempty"`
}

func (gt *Actions) AddAction(actionId string, action ResourceAction) {
	if gt.Actions == nil {
		gt.Actions = make(map[string]ResourceAction)
	}
	gt.Actions[actionId] = action
}

func (gt *Actions) POST(w http.ResponseWriter, r *http.Request) {
	if gt.Actions == nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		var actionId = requests.GetSingleQueryParameter(r, "action", "default")
		if action, ok := gt.Actions[actionId]; ok {
			action.Executer()
			w.WriteHeader(http.StatusAccepted)
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
	}
}
