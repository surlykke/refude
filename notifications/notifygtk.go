package notifications

import (
	_ "embed"
	"log"
	"time"

	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/surlykke/RefudeServices/icons"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

//go:embed gui.css
var guiCss string

func runGui() {

	gtk.Init(nil)
	if win, err := gtk.WindowNew(gtk.WINDOW_POPUP); err != nil {
		log.Fatal("Unable to create window:", err)
	} else {
		cssProvider, _ := gtk.CssProviderNew()
		cssProvider.LoadFromData(guiCss)
		screen, _ := gdk.ScreenGetDefault()
		gtk.AddProviderForScreen(screen, cssProvider, 1)
		setLayer(win)
		glib.SignalNewV("flash", glib.TYPE_NONE, 5, glib.TYPE_BOOLEAN, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_INT)
		win.Connect("flash", OnFlash)
		win.Connect("destroy", func() {
			gtk.MainQuit()
		})
		go watchNotifications(win)
		gtk.Main()
	}
}

func setLayer(win *gtk.Window) {
	layershell.InitForWindow(win)

	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_BOTTOM, true)
	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_RIGHT, true)

	layershell.SetLayer(win, layershell.LAYER_SHELL_LAYER_TOP)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_TOP, 0)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_LEFT, 0)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_RIGHT, 3)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_BOTTOM, 3)

	//layershell.SetExclusiveZone(win, 200)
}

func OnFlash(win *gtk.Window, haveFlash bool, subject string, body string, iconName string, urgency Urgency) {
	if revealer, err := win.GetChild(); err == nil && revealer != nil {
		win.Remove(revealer)
		revealer.ToWidget().Destroy()
	}
	if haveFlash {
		hbox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
		if iconfile := icons.FindIconPath(iconName, 64); iconfile != "" {
			pixbuf, _ := gdk.PixbufNewFromFileAtScale(iconfile, 42, 42, true)
			iconImage, _ := gtk.ImageNewFromPixbuf(pixbuf)
			iconImage.SetVAlign(gtk.ALIGN_START)
			hbox.PackStart(iconImage, true, true, 0)
		}
		hbox.SetName("mainbox")
		hbox.SetMarginTop(8)
		hbox.SetMarginStart(6)
		hbox.SetMarginEnd(16)
		hbox.SetMarginBottom(12)

		vbox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5)
		subjectLabel, _ := gtk.LabelNew(subject)
		subjectLabel.SetName("subject")
		subjectLabel.SetHAlign(gtk.ALIGN_START)
		vbox.PackStart(subjectLabel, true, true, 0)
		bodyLabel, _ := gtk.LabelNew(body)
		bodyLabel.SetName("body")
		if len(body) > 30 {
			bodyLabel.SetWidthChars(30)
			bodyLabel.SetLineWrap(true)
		}
		vbox.PackStart(bodyLabel, true, true, 0)
		hbox.PackStart(vbox, true, true, 0)
		win.Add(hbox)
		win.ShowAll()
	} else {
		win.Hide()
	}
}

func watchNotifications(win *gtk.Window) {
	var subscription = Notifications.Subscribe()
	updateFlash(win)
	for {
		subscription.Next()
		updateFlash(win)
	}
}

func updateFlash(win *gtk.Window) {
	var flash *Notification = nil
	var now = time.Now()
	for _, n := range Notifications.GetAll() {
		if n.Urgency == Critical {
			flash = n
			break
		} else if n.Urgency == Normal {
			if flash == nil || flash.Urgency < Normal {
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
		glib.IdleAdd(func() { win.Emit("flash", true, flash.Title, flash.Comment, flash.IconName, int(flash.Urgency)) })
		if flash.Urgency != Critical {
			var timeout = time.Time(flash.Created).Sub(time.Now()) + 50*time.Millisecond
			if flash.Urgency == Normal {
				timeout = timeout + 6*time.Second
			} else {
				timeout = timeout + 2*time.Second
			}
			time.AfterFunc(timeout, func() {
				updateFlash(win)
			})
		}
	} else {
		glib.IdleAdd(func() { win.Emit("flash", false, "", "", "", 0) })
	}
}
