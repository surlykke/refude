package options

import "github.com/jessevdk/go-flags"

type Options struct {
	NoTray          bool            `long:"no-tray" description:"Omit tray functionality"`
	NoNotifications bool            `long:"no-notifications" description:"Omit notification functionality"`
	IgnoreWinAppIds map[string]bool `long:"ignore-window" description:"Omit windows with these app-ids from search"`
}

func GetOpts() Options {
	var opts = Options{}
	if _, err := flags.Parse(&opts); err != nil {
		panic(err)
	}
	return opts
}
