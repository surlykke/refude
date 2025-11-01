// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package options

import (
	"os"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	NoNotifications bool            `long:"no-notifications" description:"Omit notification functionality"`
	NoTrayBattery   bool            `long:"no-tray-battery" description:"Dont show tray battery applet"`
	IgnoreWinAppIds map[string]bool `long:"ignore-window" description:"Omit windows with these app-ids from search"`
}

func GetOpts() Options {
	var opts = Options{}
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(0)
	}
	return opts
}
