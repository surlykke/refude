package main

import (
	"github.com/BurntSushi/xgb/composite"
	"log"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xwindow"
	"time"
	"net/http"
	"sync"
	"io/ioutil"
	"encoding/json"
	"github.com/surlykke/RefudeServices/lib"
)

const HighlightMediaType lib.MediaType = "application/vnd.org.refude.wmhighlight+json"

var (
	highlight         *Highlight
	highlightMutex    sync.Mutex
	highlightRequests chan xproto.Window
	outerWhites       []*xwindow.Window
	blacks            []*xwindow.Window
	innerWhites       []*xwindow.Window
)

type PS struct {
	//Pos and Size
	x, y, w, h int
}

type HighlightData struct {
	WindowId xproto.Window `json:",omitempty"`
}

type Highlight struct {
	HighlightData
	expires time.Time
}

func (Highlight) GetSelf() lib.StandardizedPath {
	return "/highlight"
}

func (Highlight) GetMt() lib.MediaType {
	return HighlightMediaType
}

func (Highlight) PATCH(w http.ResponseWriter, r *http.Request) {
	var hd HighlightData
	if bytes, err := ioutil.ReadAll(r.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else if err = json.Unmarshal(bytes, &hd); err != nil {
		lib.ReportUnprocessableEntity(w, err)
	} else {
		setHightlight(hd.WindowId)
		w.WriteHeader(http.StatusAccepted)
	}
}

func setHightlight(winId xproto.Window) {
	var hl = Highlight{ HighlightData{winId}, time.Now().Add(500*time.Millisecond)}
	highlightRequests <- hl.WindowId
	if hl.WindowId > 0 {
		time.AfterFunc(505*time.Millisecond, reap)
	}
	highlightMutex.Lock()
	defer highlightMutex.Unlock()
	highlight = &hl
}


func makeWindow(background uint32) *xwindow.Window {
	if w, err := xwindow.Generate(xutil); err != nil {
		panic(err)
	} else {
		w.Create(overlayWindow, 0, 0, 1, 1, xproto.CwBackPixel, background)
		return w
	}
}

func makeWindows() {
	outerWhites = make([]*xwindow.Window, 4)
	blacks = make([]*xwindow.Window, 4)
	innerWhites = make([]*xwindow.Window, 4)

	for i := 0; i < 4; i++ {
		outerWhites[i] = makeWindow(0xffffff)
		blacks[i] = makeWindow(0x0)
		innerWhites[i] = makeWindow(0xffffff)
	}
}

func drawRect(x, y, w, h int) {
	if w < 10 || h < 10 {
		return // TODO
	}

	var outerWhitePS = []PS{{x, y, w, 1},
		{x + w - 1, y, 1, h},
		{x, y + h - 1, w, 1},
		{x, y, 1, h}}

	var blackPS = []PS{{x + 1, y + 1, w - 2, 3},
		{x + w - 4, y + 1, 3, h - 2},
		{x + 1, y + h - 4, w - 2, 3},
		{x + 1, y + 1, 3, h - 2}}

	var innerWhitePS = []PS{{x + 4, y + 4, w - 8, 1},
		{x + w - 5, y + 4, 1, h - 8},
		{x + 4, y + h - 5, w - 8, 1},
		{x + 4, y + 4, 1, h - 8}}

	for i := 0; i < 4; i++ {
		outerWhites[i].MoveResize(outerWhitePS[i].x, outerWhitePS[i].y, outerWhitePS[i].w, outerWhitePS[i].h)
		outerWhites[i].Map()
		blacks[i].MoveResize(blackPS[i].x, blackPS[i].y, blackPS[i].w, blackPS[i].h)
		blacks[i].Map()
		innerWhites[i].MoveResize(innerWhitePS[i].x, innerWhitePS[i].y, innerWhitePS[i].w, innerWhitePS[i].h)
		innerWhites[i].Map()
	}
}

func getOverlayWindow() (xproto.Window, error) {
	cookie := composite.GetOverlayWindow(xgbConn, xutil.RootWin())
	if getOverlayReply, err := cookie.Reply(); err != nil {
		return 0, err
	} else {
		return getOverlayReply.OverlayWin, nil
	}
}

func highlighter() {
	for hr := range highlightRequests {
		if hr == 0 {
			for i := 0; i < 4; i++ {
				outerWhites[i].Unmap()
				blacks[i].Unmap()
				innerWhites[i].Unmap()
			}
		} else {
			if rect, err := xwindow.New(xutil, hr).DecorGeometry(); err != nil {
				log.Println(err)
			} else {
				drawRect(rect.X(), rect.Y(), rect.Width(), rect.Height())
			}
		}
	}
}


func reap() {
	var hl = getHightlight();
	if hl.WindowId > 0 && hl.expires.Before(time.Now()) {
		setHightlight(0)
	}
}

func getHightlight() *Highlight {
	highlightMutex.Lock()
	defer highlightMutex.Unlock()
	return highlight
}

func setupHighlighting() {
	makeWindows()
	highlight = &Highlight{}
	highlightRequests = make(chan xproto.Window)
	go highlighter()
}
