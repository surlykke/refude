package main

import (
	"github.com/surlykke/RefudeServices/lib/service"
	"net/http"
	"strings"
	"fmt"
	"os/exec"
	"os"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"regexp"
	"github.com/surlykke/RefudeServices/lib/ini"
	"github.com/surlykke/RefudeServices/lib/utils"
	"time"
	"io/ioutil"
	"log"
	"encoding/json"
	"github.com/surlykke/RefudeServices/lib/requestutils"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/action"
)


type launchEvent struct {
	id   string
	time int64
}

var launchEvents = make(chan launchEvent)


func Run() {
	var collected = make(chan collection)

	go CollectAndWatch(collected)
	var lastLaunched = LoadLastLaunched()

	for {
		select {
		case update := <-collected:
			fmt.Println("recieving...")
			service.RemoveAll("/applications")
			service.RemoveAll("/actions")
			for _, app := range update.applications {
				if x, ok := lastLaunched[app.Id]; ok {
					app.RelevanceHint = x
				}
				service.Map("/applications/"+app.Id, resource.MakeJsonResource(app, DesktopApplicationMediaType))

				var defaultPath = "/actions/" + app.Id
				var executer = MakeExecuter(app.Exec, app.Terminal)
				var act = action.MakeAction(app.Name, app.Comment, app.IconName, defaultPath, executer)
				service.Map(defaultPath, resource.MakeJsonResource(act, action.ActionMediaType))

				for actionId, da := range app.Actions {
					var path = "/actions/" + app.Id + "-" + actionId
					var iconName = da.IconName
					if iconName == "" {
						iconName = app.IconName
					}
					var executer = MakeExecuter(da.Exec, app.Terminal)
					var act = action.MakeAction(app.Name + ": " + da.Name, app.Comment, da.IconName, path, executer)
					service.Map(path, resource.MakeJsonResource(act, action.ActionMediaType))
				}

			}
			service.RemoveAll("/mimetypes")
			for _, mt := range update.mimetypes {
				service.Map("/mimetypes/"+mt.Id, resource.MakeJsonResource(mt, MimetypeMediaType))
			}
		case le := <-launchEvents:
			lastLaunched[le.id] = le.time
			SaveLastLaunched(lastLaunched)
		}
	}
}

func MakeExecuter(exec string, runInTerminal bool) action.Executer {
	var expandedExec = regexp.MustCompile("%[uUfF]").ReplaceAllString(exec, "")
	return func() {
		runCmd(runInTerminal, expandedExec)
	}
}

func (da *DesktopApplication) POST(w http.ResponseWriter, r *http.Request) {
	var actionName = requestutils.GetSingleQueryParameter(r, "action", "")
	var args = r.URL.Query()["arg"]
	var exec string
	if actionName == "" {
		exec = da.Exec
	} else if action, ok := da.Actions[actionName]; !ok {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		exec = action.Exec
	}

	var onlySingleArg = !(strings.Contains(exec, "%F") || strings.Contains(exec, "%U"))
	if onlySingleArg && len(args) > 1 {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		var expandedExec = regexp.MustCompile("%[uUfF]").ReplaceAllString(exec, strings.Join(args, " "))
		runCmd(da.Terminal, expandedExec)
		var copy = *da
		copy.RelevanceHint = time.Now().Unix()
		launchEvents <- launchEvent{copy.Id, copy.RelevanceHint}
		service.Map("/applications/"+ copy.Id, resource.MakeJsonResource(&copy, DesktopApplicationMediaType))
		w.WriteHeader(http.StatusAccepted)
	}
}

func runCmd(runInTerminal bool, cmdStr string) {
	var cmd *exec.Cmd
	if runInTerminal {
		var terminal, ok = os.LookupEnv("TERMINAL")
		if !ok {
			reportError(fmt.Sprintf("Trying to run %s in terminal, but env variable TERMINAL not set", cmdStr))
			return
		}
		cmdStr = fmt.Sprintf("%s -e %s", terminal, cmdStr)
	}
	var argv = strings.Fields(cmdStr)
	cmd = exec.Command(argv[0], argv[1:]...)

	cmd.Dir = xdg.Home
	cmd.Stdout = nil
	cmd.Stderr = nil


	if err := cmd.Start(); err != nil {
		reportError(fmt.Sprint(err))
		return
	}

	go cmd.Wait()
}

func (mt *Mimetype) POST(w http.ResponseWriter, r *http.Request) {
	defaultAppId := r.URL.Query()["defaultApp"]
	if len(defaultAppId) != 1 || defaultAppId[0] == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		go setDefaultApp(mt.Id, defaultAppId[0])
		w.WriteHeader(http.StatusAccepted)
	}
}

func setDefaultApp(mimetypeId string, appId string) {
	path := xdg.ConfigHome + "/mimeapps.list"

	if iniFile, err := ini.ReadIniFile(path); err != nil && !os.IsNotExist(err) {
		reportError(fmt.Sprint(err))
	} else {
		var defaultGroup = iniFile.FindGroup("Default Applications")
		if defaultGroup == nil {
			defaultGroup = &ini.Group{"Default Applications", make(map[string]string)}
			iniFile = append(iniFile, defaultGroup)
		}
		var defaultAppsS = defaultGroup.Entries[mimetypeId]
		var defaultApps = utils.Split(defaultAppsS, ";")
		defaultApps = utils.PushFront(appId, utils.Remove(defaultApps, appId))
		defaultAppsS = strings.Join(defaultApps, ";")
		defaultGroup.Entries[mimetypeId] = defaultAppsS
		if err = ini.WriteIniFile(path, iniFile); err != nil {
			reportError(fmt.Sprint(err))
		}
	}
}

func reportError(msg string) {
	log.Println(msg)
}

var lastLaunchedDir = xdg.ConfigHome + "/RefudeDesktopService"
var lastLaunchedPath = lastLaunchedDir + "/lastLaunched.json"

func LoadLastLaunched() map[string]int64 {
	var lastLaunched = make(map[string]int64)
	if bytes, err := ioutil.ReadFile(lastLaunchedPath); err != nil {
		log.Println("Error reading", lastLaunchedPath, ", ", err)
	} else if err := json.Unmarshal(bytes, &lastLaunched); err != nil {
		log.Println("Error unmarshalling lastLaunched", err)
	}
	return lastLaunched
}

func SaveLastLaunched(lastLaunched map[string]int64) {
	if bytes, err := json.Marshal(lastLaunched); err != nil {
		log.Println("Error marshalling lastLaunched", err)
	} else if err = os.MkdirAll(lastLaunchedDir, 0755); err != nil {
		log.Println("Error creating dir", lastLaunchedDir, err)
	} else if err = ioutil.WriteFile(lastLaunchedPath, bytes, 0644); err != nil {
		log.Println("Error writing lastLaunched", err)
	}
}
