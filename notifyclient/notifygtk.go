package notifyclient

import (
	_ "embed"
	"fmt"
	"log"
	"time"

	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/notifications"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

//go:embed gui.css
var guiCss string

var (
	cssProvider  *gtk.CssProvider
	screen *gdk.Screen
	win          *gtk.Window
	hbox         *gtk.Box
	iconImage    *gtk.Image
	vbox         *gtk.Box
	subjectLabel *gtk.Label
	bodyLabel    *gtk.Label
	err          error
)

func Run() {

	gtk.Init(nil)
	setup()

	glib.SignalNewV("flash", glib.TYPE_NONE, 5, glib.TYPE_BOOLEAN, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_INT)
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	go watchNotifications(win)
	gtk.Main()
}

func setup() {
	if cssProvider, err = gtk.CssProviderNew(); err != nil {
		log.Fatal("Unable to create cssProvider:", err)
	}
	cssProvider.LoadFromData(guiCss)

	if screen, err = gdk.ScreenGetDefault(); err != nil {
		log.Fatal("Unable to get screen", err)
	}

	gtk.AddProviderForScreen(screen, cssProvider, 1)
	if hbox, err = gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5); err != nil {
		log.Fatal("Unable to create hbox", err)
	}

	if win, err = gtk.WindowNew(gtk.WINDOW_POPUP); err != nil {
		log.Fatal("Unable to create window:", err)
	}

	layershell.InitForWindow(win)
	layershell.SetLayer(win, layershell.LAYER_SHELL_LAYER_TOP)
	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_BOTTOM, true)
	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_RIGHT, true)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_RIGHT, 5)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_BOTTOM, 5)


	if hbox, err = gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5); err != nil {
		log.Fatal("Unable to create hbox", err)
	}
	win.Add(hbox)

	hbox.SetName("mainbox")
	hbox.SetMarginTop(8)
	hbox.SetMarginStart(6)
	hbox.SetMarginEnd(16)
	hbox.SetMarginBottom(12)

	if iconImage, err = gtk.ImageNew(); err != nil {
		log.Fatal("Unable to create iconImage", err)
	}
	hbox.PackStart(iconImage, true, true, 0)

	if vbox, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5); err != nil {
		log.Fatal("Unable to create vbox", err)
	}
	hbox.PackStart(vbox, true, true, 0)

	if subjectLabel, err = gtk.LabelNew(""); err != nil {
		log.Fatal("Unable to create subjectLabel", err)
	}
	vbox.PackStart(subjectLabel, true, true, 0)
	subjectLabel.SetName("subject")
	subjectLabel.SetHAlign(gtk.ALIGN_START)

	if bodyLabel, err = gtk.LabelNew(""); err != nil {
		log.Fatal("Unable to to create bodyLabel", err)
	}
	bodyLabel.SetLineWrap(true)
	vbox.PackStart(bodyLabel, true, true, 0)
	bodyLabel.SetName("body")
}

func update(flash *notifications.Notification) {
	// win *gtk.Window, haveFlash bool, subject string, body string, iconName string, urgency notifications.Urgency
	if flash == nil {
		win.Hide()
	} else {
		var iconFile = ""
		if flash.IconName != "" {
			iconFile = icons.FindIconPath(flash.IconName, 64)
		}
		if iconFile != "" {
			iconImage.SetFromFile(iconFile)
		} else {
			iconImage.Clear()
		}
		subjectLabel.SetText(flash.Title)
		fmt.Println("comment len:", len(flash.Comment))
		if len(flash.Comment) > 30 {
			fmt.Println("SetMaxWidthChars")	
			bodyLabel.SetWidthChars(30)
		} else {
			fmt.Println("UnSetMaxWidthChars")	
			bodyLabel.SetWidthChars(-1)
		}
		bodyLabel.SetText(flash.Comment)
		win.ShowAll()
	}
}

// To be called in glib mainloop
func getFlash() {
	var flash *notifications.Notification = nil
	var now = time.Now()
	for _, n := range notifications.Notifications.GetAll() {
		if n.Urgency == notifications.Critical {
			flash = n
			break
		} else if n.Urgency == notifications.Normal {
			if flash == nil || flash.Urgency < notifications.Normal {
				if now.Before(time.Time(n.Created).Add(6 * time.Second)) {
					flash = n
				}
			}
		} else { /* n.Urgency == Low */
			if flash == nil && now.Before(time.Time(n.Created).Add(2*time.Second)) {
				flash = n
			}
		}
	}

	if flash != nil {
		if flash.Urgency != notifications.Critical {
			var timeout = time.Time(flash.Created).Sub(time.Now()) + 50*time.Millisecond
			if flash.Urgency == notifications.Normal {
				timeout = timeout + 6*time.Second
			} else {
				timeout = timeout + 2*time.Second
			}
			time.AfterFunc(timeout, func() {
				glib.IdleAdd(getFlash)
			})
		}
	} 
	update(flash)

}

func watchNotifications(win *gtk.Window) {
	var subscription = notifications.Notifications.Subscribe()
	glib.IdleAdd(getFlash)
	for {
		subscription.Next()
		glib.IdleAdd(getFlash)
	}
}
