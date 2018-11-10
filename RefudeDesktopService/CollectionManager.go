// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"encoding/json"
	"fmt"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
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
			resourceHandler.RemoveAll("/applications")
			resourceHandler.RemoveAll("/actions")
			for _, app := range update.applications {
				if x, ok := lastLaunched[app.Id]; ok {
					app.RelevanceHint = x
				}

				resourceHandler.Map(app)
			}
			resourceHandler.RemoveAll("/mimetypes")
			for _, mt := range update.mimetypes {
				resourceHandler.Map(mt)
				for _,alias := range mt.Aliases {
					resourceHandler.MapTo(resource.Standardizef("/mimetypes/%s", alias), mt)
				}
			}
		case le := <-launchEvents:
			lastLaunched[le.id] = le.time
			SaveLastLaunched(lastLaunched)
		}
	}
}

func MakeExecuter(exec string, runInTerminal bool) resource.Executer {
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
		xdg.RunCmd(argv)
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
