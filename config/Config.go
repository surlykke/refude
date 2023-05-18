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
	Enabled              bool
	BatteryNotifications bool
	Placement            []Placement
}

type LauncherConf struct {
	Placement []Placement
}

type Placement struct {
	Screen      string
	Corner      uint8
	CornerdistX uint
	CornerdistY uint
	Width       uint
	Height      uint
}

type Conf struct {
	Notifications NotificationConf
	Launcher      LauncherConf
}

var Notifications = NotificationConf{
	Enabled:              true,
	BatteryNotifications: true,
}

var Launcher = LauncherConf{}

func init() {
	var tmp Conf
	var bytes []byte
	var err error

	if _, err = os.Stat(configFile); err != nil && os.IsNotExist(err) {
		log.Info(configFile, "not found")
	} else if bytes, err = os.ReadFile(configFile); err != nil {
		log.Warn("Unable to read", configFile, err)
	} else if err = yaml.Unmarshal(bytes, &tmp); err != nil {
		log.Warn("Unable to parse", configFile, err)
	} else {
		Notifications = tmp.Notifications
		Launcher = tmp.Launcher
	}

	fmt.Println("After config init")
	fmt.Println("Notifications:", Notifications)
	fmt.Println("Launcher:", Launcher)
	fmt.Println("yaml file was:", string(bytes))
}
