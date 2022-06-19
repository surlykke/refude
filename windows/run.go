package windows

import (
	"github.com/surlykke/RefudeServices/windows/x11"
)

func Run() {
	var proxy = x11.MakeProxy()
	x11.SubscribeToEvents(proxy)
	updateWindowList(proxy)
	for {
		event, _ := x11.NextEvent(proxy)
		if event == x11.DesktopStacking {
			updateWindowList(proxy)
		} else if event == x11.ActiveWindow {
			if activeWindow, err := x11.GetActiveWindow(proxy); err == nil {
				recentlyFocusedWindows.add(XWin(activeWindow))	
			}
		}
	}
}

func updateWindowList(p x11.Proxy) {
	var wIds = x11.GetStack(p)
	var xWins = make([]XWin, len(wIds), len(wIds))
	for i := 0; i < len(wIds); i++ {
		xWins[i] = XWin(wIds[i])
	}
	Windows.ReplaceWith(xWins)
}

