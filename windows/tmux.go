package windows

import (
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

/*func ServeHTTP(w http.ResponseWriter, r *http.Request) {

}*/

type TmuxPane struct {
	PaneId           string
	CurrentCommand   string
	CurrentDirectory string
	WindowId         string
	XWinId           uint32
}

type TmuxPaneShort struct {
	PaneId           string
	WindowId         string
	CurrentCommand   string
	CurrentDirectory string
}

func (t *TmuxPane) Id() string {
	return t.PaneId
}

func (t *TmuxPane) Presentation() (title string, comment string, iconUrl link.Href, profile string) {
	return t.CurrentCommand + ":" + t.CurrentDirectory, "", "", "tmuxpane"
}

func (t *TmuxPane) Links(self, term string) link.List {
	return link.List{
		link.Make(self, "Focus", "", relation.DefaultAction),
	}
}

func (t *TmuxPane) DoPost(w http.ResponseWriter, r *http.Request) {
	var action = requests.GetSingleQueryParameter(r, "action", "")
	if action == "" {
		if err := exec.Command("tmux", "select-pane", "-t", "%"+t.PaneId).Run(); err != nil {
			respond.ServerError(w, err)
		} else if err = exec.Command("tmux", "select-window", "-t", t.WindowId).Run(); err != nil {
			respond.ServerError(w, err)
		} else {
			// FIXME
			/*x11.ProxyMutex.Lock()
			x11.RaiseAndFocusWindow(x11.SynchronizedProxy, t.XWinId)
			x11.ProxyMutex.Unlock()*/
			respond.Accepted(w)
		}

	} else {
		respond.NotFound(w)
	}
}

func collectShortPaneMap() map[string]TmuxPaneShort {
	var panes = make(map[string]TmuxPaneShort, 20)
	// We assume #{pane_current_command} does not contain '/'
	for _, line := range tmux("list-panes", "#{pane_id}", "#{window_id}", "#{pane_current_command}", "#{pane_current_path}") {
		if len(line) >= 4 && strings.HasPrefix(line[0], "%") {
			panes[line[0]] = TmuxPaneShort{
				PaneId:           line[0][1:],
				WindowId:         line[1],
				CurrentCommand:   line[2],
				CurrentDirectory: line[3],
			}
		}
	}
	return panes
}

// Returns map keyed by window ids, with values owning session id
func collectWindowSessionMap() map[string]string {
	var result = make(map[string]string, 10)
	for _, values := range tmux("list-windows", "#{window_id}", "#{session_id}") {
		if len(values) == 2 {
			result[values[0]] = values[1]
		}
	}
	return result
}

func collectSessionClientMap() map[string]string {
	var result = make(map[string]string, 10)

	for _, values := range tmux("list-sessions", "#{session_id}", "#{session_attached_list}") {
		if len(values) == 2 {
			result[values[0]] = values[1]
		}
	}
	return result
}

func collectClientPidMap() map[string]string {
	var result = make(map[string]string, 10)

	for _, values := range tmux("list-clients", "#{client_tty}", "#{client_pid}") {
		if len(values) == 2 {
			result[values[0]] = values[1]
		}
	}
	return result

}

func getParentPid(pid string) string {
	var procfile = "/proc/" + pid + "/status"
	if bytes, err := ioutil.ReadFile(procfile); err != nil {
		log.Warn("Error reading", procfile, err)
		return ""
	} else {
		var s = string(bytes)
		if ppidPos := strings.Index(s, "PPid:"); ppidPos == -1 {
			return ""
		} else {
			var i = ppidPos + 5 + 1 // 'PPid:' is followed by a tab
			var j = i
			for ; j < len(s) && '0' <= s[j] && s[j] <= '9'; j++ {
			}
			if j > i {
				return s[i:j]
			} else {
				return ""
			}
		}
	}

}

func collectPanes() map[string]*TmuxPane {
	var shortPanesMap = collectShortPaneMap()
	var windowSessionMap = collectWindowSessionMap()
	var sessionClientMap = collectSessionClientMap()
	var clientPidMap = collectClientPidMap()

	var panes = make(map[string]*TmuxPane, len(shortPanesMap))

	for _, shortPane := range shortPanesMap {
		if sessionId, ok := windowSessionMap[shortPane.WindowId]; ok {
			if clientId, ok := sessionClientMap[sessionId]; ok {
				if clientPid, ok := clientPidMap[clientId]; ok {
					for pid := clientPid; pid != ""; pid = getParentPid(pid) {
						panes[shortPane.PaneId] = &TmuxPane{
							PaneId:           shortPane.PaneId,
							CurrentCommand:   shortPane.CurrentCommand,
							CurrentDirectory: shortPane.CurrentDirectory,
							WindowId:         shortPane.WindowId,
						}
						break
					}
				}
			}
		}
	}
	return panes
}

func SearchTmuxPanes(term string) link.List {
	var links = make(link.List, 0, 20)
	for _, pane := range collectShortPaneMap() {
		var title = pane.CurrentCommand + ":" + pane.CurrentDirectory
		if rnk := searchutils.Match(term, title, "tmux", "pane"); rnk > -1 {
			links = append(links, link.MakeRanked("/tmux/"+pane.PaneId, title, "", "tmuxpane", rnk))
		}
	}

	return links
}

// Returns lines retrieved from tmux command as []string
func tmux(cmd string, formatElements ...string) [][]string {
	// Used as a separator. Hopefully sufficiently random that we wont see it
	// as part of a command or directory name. If we do, things will go south
	const separator = "2060230513"

	var args []string
	if cmd == "list-panes" || cmd == "list-windows" {
		args = []string{cmd, "-a", "-F", strings.Join(formatElements, separator)}
	} else {
		args = []string{cmd, "-F", strings.Join(formatElements, separator)}
	}

	if res, err := exec.Command("tmux", args...).Output(); err != nil {
		return [][]string{}
	} else {
		var lines = strings.Split(string(res), "\n")
		var result = make([][]string, 0, len(lines))

		for i := 0; i < len(lines); i++ {
			var values = strings.Split(strings.TrimSpace(lines[i]), separator)
			if len(values) == len(formatElements) {
				result = append(result, values)
			}
		}

		return result
	}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/tmux/") {
		var panes = collectPanes()
		if pane, ok := panes[r.URL.Path[6:]]; ok {
			resource.ServeResource[string](w, r, r.URL.Path, pane)
			return
		}
	}
	respond.NotFound(w)
}
