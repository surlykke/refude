package main

import (
	"github.com/surlykke/RefudeServices/lib/xdg"
	"io/ioutil"
	"encoding/json"
	"log"
	"os"
)

var lastLaunched map[string]int64

var lastLaunchedDir = xdg.ConfigHome + "/RefudeDesktopService"
var lastLaunchedPath = lastLaunchedDir + "/lastLaunched.json"


func LoadLastLaunched() {
	lastLaunched = make(map[string]int64)
	if bytes, err := ioutil.ReadFile(lastLaunchedPath); err != nil {
		log.Println("Error reading", lastLaunchedPath, ", ", err)
	} else if err := json.Unmarshal(bytes, &lastLaunched); err != nil {
		log.Println("Error unmarshalling lastLaunched", err)
	}
}


func SaveLastLaunched() {
	if bytes, err := json.Marshal(lastLaunched); err != nil {
		log.Println("Error marshalling lastLaunched", err)
	} else if err = os.MkdirAll(lastLaunchedDir, 0755); err != nil {
		log.Println("Error creating dir", lastLaunchedDir, err)
	} else if err = ioutil.WriteFile(lastLaunchedPath, bytes, 0644); err != nil {
		log.Println("Error writing lastLaunched", err)
	}
}

