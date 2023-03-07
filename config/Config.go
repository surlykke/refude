package config

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

//go:embed defaultConfig.ini
var defaultConfig string

var configDir = xdg.ConfigHome + "/refude"
var configFile = configDir + "/config"

type Config  struct {
	noNotificationServer bool 
	noBatteryIcon bool 
	noBatteryNotifications bool 
}


var conf Config

func Read() {
	fmt.Println("Read...")
	os.MkdirAll(configDir, 0770)
	if _, err := os.Stat(configFile); err != nil {
		fmt.Println("stat err:", err)
		if os.IsNotExist(err) {
			ioutil.WriteFile(configFile, []byte(defaultConfig), 0660)
		} else {
			log.Warn("Could not stat", configFile, err)
			return
		}
	}
	
	if iniFile, err := xdg.ReadIniFile(configFile); err != nil {
		log.Warn("Could not read", configFile, err)
	} else if general := iniFile.FindGroup("General"); general == nil {
		log.Warn("Found no group 'General' in", configFile)
		return 
	} else {
		conf.noNotificationServer = "true" == general.Entries["noNotificationServer"]
		conf.noBatteryIcon = "true" == general.Entries["noBatteryIcon"]
		conf.noBatteryNotifications = "true" == general.Entries["noBatteryNotifications"]
	}
}

func NoNotificationServer() bool {
	return conf.noNotificationServer
}

func NoBatteryIcon() bool {
	return conf.noBatteryIcon
}

func NoBatteryNotifications() bool {
	return conf.noBatteryNotifications
}
