package windows 

import (
	"bytes"
	"encoding/json"
	"os/exec"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

/*func ServeHTTP(w http.ResponseWriter, r *http.Request) {

}*/

type TmuxPane struct {
	Id               string
	CurrentCommand   string
	CurrentDirectory string
	WindowId         string
	SessionId        string
	ClientId         string
	ClientPid        uint32
	XWindowId        uint32
}


func SearchTmuxPanes(term string) link.List {
	var buf = bytes.NewBuffer(make([]byte, 0, 100))
	buf.WriteByte('{')
	buf.Write(get("tmux", "list-panes", "-a", "-F", `"#{pane_id}":"#{pane_current_command}:#{pane_current_path}"`))
	buf.WriteByte('}')

	var res = buf.Bytes()
	var lastPos = len(res) - 2
	for i, b := range res {
		if b == '\n' && i != lastPos {
			res[i] = ','
		}
	}

	var m = make(map[string]string)
	if err := json.Unmarshal(res, &m); err != nil {
		return link.List{}
	} else {
		var list = make(link.List, 0, len(m))
		for id, title := range m {
			if rnk := searchutils.Match(term, title); rnk > -1 {
				list = append(list, link.MakeRanked("/tmuxpane/" + id, title, "", "tmuxpane", rnk))
			}
		}
		return list;
	}
}




func get(cmd string, args ...string) []byte {
	if res, err := exec.Command(cmd, args...).Output(); err != nil {
		return []byte{}
	} else {
		return res
	}
}


/*func getPanes() []tmuxPane {
	var buf = bytes.NewBuffer(make([]byte, 0, 100))
	buf.WriteByte('{')
	buf.Write(get("tmux", "list-clients", "-F", `"#{client_tty}":"#{client_pid}"`))
	buf.Write(get("tmux", "list-sessions", "-F", `"#{session_id}":"#{session_attached_list}"`))
	buf.Write(get("tmux", "list-windows", "-a", "-F", `"#{window_id}":"#{session_id}"`))
	buf.Write(get("tmux", "list-panes", "-a", "-F", `"#{pane_id}":"#{window_id}"`))
	buf.WriteByte('}')
	var res = buf.Bytes()
	var lastPos = len(res) - 2
	for i, b := range res {
		if b == '\n' && i != lastPos {
			res[i] = ','
		}
	}
	return buf.Bytes()
}*/
