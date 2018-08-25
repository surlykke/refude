// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"net/http"
	"strings"
	"fmt"
	"os"
	"github.com/surlykke/RefudeServices/lib"
	"regexp"
	"time"
	"io/ioutil"
	"log"
	"encoding/json"
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
			resources.RemoveAll("/applications")
			resources.RemoveAll("/actions")
			for _, app := range update.applications {
				if x, ok := lastLaunched[app.Id]; ok {
					app.RelevanceHint = x
				}

				if ! app.NoDisplay {
					var defaultPath = lib.Standardizef("/actions/%s", app.Id)
					var executer = MakeExecuter(app.Exec, app.Terminal)
					var act = lib.MakeAction(defaultPath, app.Name, app.Comment, app.IconName, executer)
					act.RelevanceHint = app.RelevanceHint
					lib.Relate(&app.AbstractResource, &act.AbstractResource)
					resources.Map(act)
				}
				resources.Map(app)

			}
			resources.RemoveAll("/mimetypes")
			for _, mt := range update.mimetypes {
				resources.Map(mt)
				for _,alias := range mt.Aliases {
					resources.MapTo(lib.Standardizef("/mimetypes/%s", alias), mt)
				}
			}
		case le := <-launchEvents:
			lastLaunched[le.id] = le.time
			SaveLastLaunched(lastLaunched)
		}
	}
}

func MakeExecuter(exec string, runInTerminal bool) lib.Executer {
	var expandedExec = regexp.MustCompile("%[uUfF]").ReplaceAllString(exec, "")
	var argv []string
	if runInTerminal {
		var terminal, ok = os.LookupEnv("TERMINAL")
		if !ok {
			reportError(fmt.Sprintf("Trying to make executer for %s in terminal, but env variable TERMINAL not set", exec))
			return func() {}
		}
		argv = []string{terminal, "-e", "'" + strings.TrimSpace(strings.Replace(expandedExec, "'", "'\\''", -1)) + "'"}
	} else {
		argv = strings.Fields(expandedExec)
	}

	return func() {
		lib.RunCmd(argv)
	}
}

func (da *DesktopApplication) POST(w http.ResponseWriter, r *http.Request) {
	fmt.Println("In post")
	var actionName = lib.GetSingleQueryParameter(r, "action", "")
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
		var argv []string
		var argsReg = regexp.MustCompile("%[uUfF]");
		if da.Terminal {
			var terminal, ok = os.LookupEnv("TERMINAL")
			if !ok {
				reportError(fmt.Sprintf("Trying to run %s in terminal, but env variable TERMINAL not set", exec))
				return
			}
			var arglist = []string{}
			for _, arg := range args {
				arglist = append(arglist, "'"+strings.Replace(arg, "'", "'\\''", -1)+"'")
			}
			var argListS = strings.Join(arglist, " ");
			var cmd = argsReg.ReplaceAllString(exec, argListS)
			fmt.Println("Run in terminal with cmd:", cmd)
			argv = []string{terminal, "-e", cmd}
		} else {
			var fields = strings.Fields(exec)
			for _, field := range fields {
				if argsReg.MatchString(field) {
					argv = append(argv, args...)
				} else {
					argv = append(argv, field)
				}
			}
		}

		lib.RunCmd(argv)

		var copy = *da
		copy.Relates = make(map[lib.MediaType][]lib.StandardizedPath)
		for mt, urls := range da.Relates {
			copy.Relates[mt] = urls
		}
		copy.RelevanceHint = time.Now().Unix()
		launchEvents <- launchEvent{copy.Id, copy.RelevanceHint}
		resources.Map(&copy)
		w.WriteHeader(http.StatusAccepted)
	}
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
	path := lib.ConfigHome + "/mimeapps.list"

	if iniFile, err := lib.ReadIniFile(path); err != nil && !os.IsNotExist(err) {
		reportError(fmt.Sprint(err))
	} else {
		var defaultGroup = iniFile.FindGroup("Default Applications")
		if defaultGroup == nil {
			defaultGroup = &lib.Group{"Default Applications", make(map[string]string)}
			iniFile = append(iniFile, defaultGroup)
		}
		var defaultAppsS = defaultGroup.Entries[mimetypeId]
		var defaultApps = lib.Split(defaultAppsS, ";")
		defaultApps = lib.PushFront(appId, lib.Remove(defaultApps, appId))
		defaultAppsS = strings.Join(defaultApps, ";")
		defaultGroup.Entries[mimetypeId] = defaultAppsS
		if err = lib.WriteIniFile(path, iniFile); err != nil {
			reportError(fmt.Sprint(err))
		}
	}
}

func reportError(msg string) {
	log.Println(msg)
}

var lastLaunchedDir = lib.ConfigHome + "/RefudeDesktopService"
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
