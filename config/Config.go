package config

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"gopkg.in/yaml.v2"
)

//go:embed defaultConfig.yaml
var defaultConfig string

var configDir = xdg.ConfigHome + "/refude"
var configFile = configDir + "/config.yaml"

type NotificationConf struct {
	Enabled             bool
	BatteryNotifications bool
	Placement            []struct {
		Screen      string
		Corner      uint8
		CornerDistX int
		CornerDistY int
	}
}

type Conf struct {
	Notifications NotificationConf 
}

var Notifications = NotificationConf{
	Enabled: true,
	BatteryNotifications: true,
}

func init() {
	var tmp Conf 
	var bytes []byte
	var err error

	if _, err = os.Stat(configFile); err != nil && os.IsNotExist(err) {
		log.Info(configFile, "not found")
	} else if bytes, err = os.ReadFile(configFile); err != nil {
		log.Warn("Unable to read", configFile, err)
	} else if err = yaml.UnmarshalStrict(bytes, &tmp); err != nil {
		log.Warn("Unable to parse", configFile, err)
	} else {
		Notifications = tmp.Notifications
	}

	fmt.Println("After config init, Notifications:", Notifications, ", yaml file was:", string(bytes))
}
