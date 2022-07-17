// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package x11

/**
 * Interfacing with X11 happens through this file
 */

/*
#cgo LDFLAGS: -lX11 -lXrandr
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <X11/extensions/Xrandr.h>

// Cant use 'type' in Go, hence...
inline int getType(XEvent* e) { return e->type; }

// Using C macros in Go seems tricky, so..
inline int ds(Display* d) { return DefaultScreen(d); }
inline Window rw(Display *d, int screen) { return RootWindow(d, screen); }
inline unsigned long gp(XImage* img, int x, int y) {return XGetPixel(img, x, y); }

// Accessing a field inside a union inside a struct from Go is messy. Hence these helpers
inline XConfigureEvent* xconfigure(XEvent* e) { return &(e->xconfigure); }
inline XPropertyEvent* xproperty(XEvent* e) { return &(e->xproperty); }
// Converting sequences unsigned chars to byte or long. Most easily done in C, so..
const unsigned long sizeOfLong = sizeof(long);
inline char getByte(unsigned char* data, int index) { return ((char*)data)[index]; }
inline long getLong(unsigned char* data, int index) { return ((long*)data)[index]; }

XEvent createClientMessage32(Window window, Atom message_type, long l0, long l1, long l2, long l3, long l4) {
	XEvent event;
	memset(&event, 0, sizeof(XEvent));
	event.xclient.type = ClientMessage;
	event.xclient.serial = 0;
	event.xclient.send_event = 1;
	event.xclient.message_type = message_type;
	event.xclient.window = window;
	event.xclient.format = 32;
	event.xclient.data.l[0] = l0;
	event.xclient.data.l[1] = l1;
	event.xclient.data.l[2] = l2;
	event.xclient.data.l[3] = l3;
	event.xclient.data.l[4] = l4;
	return event;
}

XEvent createConfigureMessage32(Window window, Window eventWindow, int x, int y, int width, int height) {
	XEvent event;
	memset(&event, 0, sizeof(XEvent));
	event.xconfigure.window = window;
	event.xconfigure.event = eventWindow;
	event.xconfigure.send_event = 1;
	event.xconfigure.x = x;
	event.xconfigure.y = y;
	event.xconfigure.width = width;
	event.xconfigure.height = height;
	return event;
}

void setProp32(Display *d, Window w, Atom prop, Atom type, unsigned int val) {
	unsigned int arr[1];
	arr[0] = val;
	printf("Call XChangeProperty");
	XChangeProperty(d, w, prop, type, 32, PropModeReplace, (unsigned char*)arr, 1);
}


int forgiving_X_error_handler(Display *d, XErrorEvent *e)
{
	char errorMsg[80];
	XGetErrorText(d, e->error_code, errorMsg, 80);
	printf("Got error: %s\n", errorMsg);
	return 0;
}


*/
import "C"
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"sync"
	"time"
	"unsafe"

	"github.com/surlykke/RefudeServices/lib/log"
)

type MonitorData struct {
	X, Y     int32
	W, H     uint32
	Wmm, Hmm uint32
	Name     string
}

type Event uint8

const (
	DesktopGeometry Event = iota
	DesktopStacking
	ActiveWindow
	WindowTitle
	WindowIconName
	WindowGeometry
	WindowSt
)

func (e Event) String() string {
	switch e {
	case DesktopGeometry:
		return "DesktopGeometry"
	case DesktopStacking:
		return "DesktopStacking"
	case WindowTitle:
		return "WindowTitle"
	case WindowIconName:
		return "WindowIconName"
	case WindowGeometry:
		return "WindowGeometry"
	case WindowSt:
		return "WindowSt"
	default:
		return ""
	}
}

type WindowStateMask uint16

const (
	MODAL WindowStateMask = 1 << iota
	STICKY
	MAXIMIZED_VERT
	MAXIMIZED_HORZ
	SHADED
	SKIP_TASKBAR
	SKIP_PAGER
	HIDDEN
	FULLSCREEN
	ABOVE
	BELOW
	DEMANDS_ATTENTION
)

func (wsm WindowStateMask) Is(other WindowStateMask) bool {
	return wsm&other == other
}

func (wsm WindowStateMask) MarshalJSON() ([]byte, error) {
	var list = make([]string, 0, 14)
	if wsm&MODAL > 0 {
		list = append(list, "MODAL")
	}
	if wsm&STICKY > 0 {
		list = append(list, "STICKY")
	}
	if wsm&MAXIMIZED_VERT > 0 {
		list = append(list, "MAXIMIZED_VERT")
	}
	if wsm&MAXIMIZED_HORZ > 0 {
		list = append(list, "MAXIMIZED_HORZ")
	}
	if wsm&SHADED > 0 {
		list = append(list, "SHADED")
	}
	if wsm&SKIP_TASKBAR > 0 {
		list = append(list, "SKIP_TASKBAR")
	}
	if wsm&SKIP_PAGER > 0 {
		list = append(list, "SKIP_PAGER")
	}
	if wsm&HIDDEN > 0 {
		list = append(list, "HIDDEN")
	}
	if wsm&FULLSCREEN > 0 {
		list = append(list, "FULLSCREEN")
	}
	if wsm&ABOVE > 0 {
		list = append(list, "ABOVE")
	}
	if wsm&BELOW > 0 {
		list = append(list, "BELOW")
	}
	if wsm&DEMANDS_ATTENTION > 0 {
		list = append(list, "DEMANDS_ATTENTION")
	}
	return json.Marshal(list)
}

type Proxy struct {
	*sync.Mutex
	disp       *C.Display
	rootWindow C.Window
}

func MakeProxy() Proxy {
	var disp = C.XOpenDisplay(nil)
	var defaultScreen = C.ds(disp)
	var rootWindow = C.rw(disp, defaultScreen)
	return Proxy{
		Mutex: &sync.Mutex{},
		disp:       disp,
		rootWindow: rootWindow,
	}
}

var commonProxy = MakeProxy()

const redirectAndNotifyMask = C.SubstructureRedirectMask | C.SubstructureNotifyMask

func SubscribeToEvents(p Proxy) {
	C.XSelectInput(p.disp, p.rootWindow, C.PropertyChangeMask)
	C.XRRSelectInput(p.disp, p.rootWindow, C.RRScreenChangeNotifyMask)
}

func SubscribeToWindowEvents(p Proxy, wId uint32) {
	C.XSelectInput(p.disp, C.Window(wId), C.PropertyChangeMask)
}

func NextEvent(p Proxy) (Event, uint32) {
	var event C.XEvent
	for {
		if err := CheckError(C.XNextEvent(p.disp, &event)); err != nil {
			log.Warn("Error retrieving event from X11:", err)
			time.Sleep(100 * time.Millisecond) // To avoid spamming
		} else if C.getType(&event) == C.PropertyNotify {
			var xproperty = C.xproperty(&event)
			if xproperty.atom == _NET_CLIENT_LIST_STACKING {
				return DesktopStacking, 0
			} else if xproperty.atom == _NET_DESKTOP_GEOMETRY {
				return DesktopGeometry, 0
			} else if xproperty.atom == _NET_WM_VISIBLE_NAME || xproperty.atom == _NET_WM_NAME || xproperty.atom == _WM_NAME {
				return WindowTitle, uint32(xproperty.window)
			} else if xproperty.atom == _NET_WM_VISIBLE_ICON_NAME || xproperty.atom == _NET_WM_ICON_NAME || xproperty.atom == _NET_WM_ICON {
				return WindowIconName, uint32(xproperty.window)
			} else if xproperty.atom == _NET_WM_STATE {
				return WindowSt, uint32(xproperty.window)
			} else if xproperty.atom == _NET_ACTIVE_WINDOW {
				return ActiveWindow, uint32(xproperty.window)
			}
		} else if C.getType(&event) == C.ConfigureNotify {
			var xconfigure = C.xconfigure(&event)
			return WindowGeometry, uint32(xconfigure.window)
		}
	}
}

func GetMonitorDataList(p Proxy) []*MonitorData {
	var num C.int
	xrrmonitorsPtr := C.XRRGetMonitors(p.disp, p.rootWindow, 1, &num)
	xrrmonitorsArr := (*[1 << 30]C.XRRMonitorInfo)(unsafe.Pointer(xrrmonitorsPtr))
	var monitors = make([]*MonitorData, num, num)
	var bound int = int(num)
	for i := 0; i < bound; i++ {
		monitors[i] = &MonitorData{
			X:    int32(xrrmonitorsArr[i].x),
			Y:    int32(xrrmonitorsArr[i].y),
			W:    uint32(xrrmonitorsArr[i].width),
			H:    uint32(xrrmonitorsArr[i].height),
			Wmm:  uint32(xrrmonitorsArr[i].mwidth),
			Hmm:  uint32(xrrmonitorsArr[i].mheight),
			Name: getAtomName(p.disp, xrrmonitorsArr[i].name),
		}
	}
	C.XRRFreeMonitors(xrrmonitorsPtr)
	return monitors
}

func GetParent(p Proxy, wId uint32) (uint32, error) {
	var root_return C.ulong
	var parent_return C.ulong
	var children_return *C.ulong
	var nchildren_return C.uint
	for {
		if C.XQueryTree(p.disp, C.ulong(wId), &root_return, &parent_return, &children_return, &nchildren_return) == 0 {
			return 0, errors.New("Error from XQueryTree")
		} else {
			if children_return != nil {
				C.XFree(unsafe.Pointer(children_return))
			}
			if parent_return == p.rootWindow {
				return wId, nil
			} else {
				wId = uint32(parent_return)
			}
		}
	}
}

func GetGeometry(p Proxy, wId uint32) (int32, int32, uint32, uint32, error) {
	return getGeometry(p.disp, wId)
}

func SetTransparent(p Proxy, wId uint32, opacity uint32) {
	C.XChangeProperty(p.disp, C.Window(wId), _NET_WM_WINDOW_OPACITY, 6, 32, C.PropModeReplace, (*C.uchar)(unsafe.Pointer(&opacity)), 1)
	C.XFlush(p.disp)
}

func SetOpaque(p Proxy, wId uint32) {
	C.XDeleteProperty(p.disp, C.Window(wId), _NET_WM_WINDOW_OPACITY)
	C.XFlush(p.disp)
}

// Returns wIds of current windows, stack order, bottom up
func GetStack(p Proxy) []uint32 {
	if stack, err := getUint32s(p.disp, p.rootWindow, _NET_CLIENT_LIST_STACKING); err != nil {
		log.Warn("Unable to get stack:", err)
		return []uint32{}
	} else {
		return stack
	}
}

func GetActiveWindow(p Proxy) (uint32, error) {
	if activeWindowList, err := getUint32s(p.disp, p.rootWindow, _NET_ACTIVE_WINDOW); err != nil {
		return 0, err
	} else if len(activeWindowList) != 1 {
		return 0, errors.New("Len of activeWindowList <> 1")
	} else {
		return activeWindowList[0], nil 
	}
}


func GetName(p Proxy, wId uint32) (string, error) {
	if bytes, err := getBytes(p.disp, C.Window(wId), _NET_WM_VISIBLE_NAME); err == nil {
		return string(bytes), nil
	} else if bytes, err = getBytes(p.disp, C.Window(wId), _NET_WM_NAME); err == nil {
		return string(bytes), nil
	} else if bytes, err = getBytes(p.disp, C.Window(wId), _WM_NAME); err == nil {
		return string(bytes), nil
	} else {
		return "", errors.New("Neither '_NET_WM_VISIBLE_NAME', '_NET_WM_NAME' nor 'WM_NAME' set")
	}
}

func GetIcon(p Proxy, wId uint32) ([]uint32, error) {
	return getUint32s(p.disp, C.ulong(wId), _NET_WM_ICON)
}

func GetStates(p Proxy, wId uint32) WindowStateMask {
	var state WindowStateMask = 0
	if atoms, err := getUint32s(p.disp, C.ulong(wId), _NET_WM_STATE); err != nil {
		return 0
	} else {
		for _, atom := range atoms {
			switch C.ulong(atom) {
			case _NET_WM_STATE_MODAL:
				state |= MODAL
			case _NET_WM_STATE_STICKY:
				state |= STICKY
			case _NET_WM_STATE_MAXIMIZED_VERT:
				state |= MAXIMIZED_VERT
			case _NET_WM_STATE_MAXIMIZED_HORZ:
				state |= MAXIMIZED_HORZ
			case _NET_WM_STATE_SHADED:
				state |= SHADED
			case _NET_WM_STATE_SKIP_TASKBAR:
				state |= SKIP_TASKBAR
			case _NET_WM_STATE_SKIP_PAGER:
				state |= SKIP_PAGER
			case _NET_WM_STATE_HIDDEN:
				state |= HIDDEN
			case _NET_WM_STATE_FULLSCREEN:
				state |= FULLSCREEN
			case _NET_WM_STATE_ABOVE:
				state |= ABOVE
			case _NET_WM_STATE_BELOW:
				state |= BELOW
			case _NET_WM_STATE_DEMANDS_ATTENTION:
				state |= DEMANDS_ATTENTION
			}
		}
		return state
	}
}

func GetPid(p Proxy, wId uint32) (uint32, error) {
	if pidList, err := getUint32s(p.disp, C.ulong(wId), _NET_WM_PID); err != nil {
		return 0, err
	} else if len(pidList) != 1 {
		return 0, errors.New("Ambigous pid for window")
	} else {
		return pidList[0], nil
	}
}

func AddStates(p Proxy, wId uint32, states WindowStateMask) {
	updateState(p.disp, p.rootWindow, C.Window(wId), states, 1)
}

func RemoveStates(p Proxy, wId uint32, states WindowStateMask) {
	updateState(p.disp, p.rootWindow, C.Window(wId), states, 0)
}

func SetBounds(p Proxy, wId uint32, x int32, y int32, w uint32, h uint32) {
	C.XMoveResizeWindow(p.disp, C.Window(wId), C.int(x), C.int(y), C.uint(w), C.uint(h))
	C.XFlush(p.disp)
}

func Resize(p Proxy, wId uint32, w uint32, h uint32) {
	C.XResizeWindow(p.disp, C.Window(wId), C.uint(w), C.uint(h))
	C.XFlush(p.disp)
}

func RaiseWindow(p Proxy, wId uint32) {
	var event = C.createClientMessage32(C.Window(wId), _NET_RESTACK_WINDOW, 2, 0, 0, 0, 0)
	C.XSendEvent(p.disp, p.rootWindow, 0, redirectAndNotifyMask, &event)
	C.XFlush(p.disp)
}

func RaiseAndFocusWindow(p Proxy, wId uint32) {
	var event = C.createClientMessage32(C.Window(wId), _NET_ACTIVE_WINDOW, 2, 0, 0, 0, 0)
	C.XSendEvent(p.disp, p.rootWindow, 0, redirectAndNotifyMask, &event)
	C.XFlush(p.disp)
}

func CloseWindow(p Proxy, wId uint32) {
	var event = C.createClientMessage32(C.Window(wId), _NET_CLOSE_WINDOW, 2, 0, 0, 0, 0)
	C.XSendEvent(p.disp, p.rootWindow, 0, redirectAndNotifyMask, &event)
	C.XFlush(p.disp)

}

func GetScreenshotAsPng(p Proxy, wId uint32, downscale uint8) ([]byte, error) {
	var ximage, w, h = getScreenShotAsXImage(p.disp, wId)

	if ximage == nil {
		return nil, fmt.Errorf("Unable to retrieve screendump for %d", wId)
	}

	pngData := image.NewRGBA(image.Rect(0, 0, int(w/uint32(downscale)), int(h/uint32(downscale))))

	for i := 0; i < int(w); i = i + int(downscale) {
		for j := 0; j < int(h); j = j + int(downscale) {
			var pixel = C.gp(ximage, C.int(i), C.int(j))
			pngData.Set(i/int(downscale), j/int(downscale), color.RGBA{R: uint8((pixel >> 16) & 255), G: uint8((pixel >> 8) & 255), B: uint8(pixel & 255), A: 255})
		}
	}

	buf := &bytes.Buffer{}
	var encoder = &png.Encoder{
		CompressionLevel: png.NoCompression,
	}
	if err := encoder.Encode(buf, pngData); err == nil {
		return buf.Bytes(), nil
	} else {
		return nil, err
	}
}

func updateState(disp *C.Display, rootWindow C.Window, win C.Window, state WindowStateMask, addRemove C.long) {
	if state&MODAL > 0 {
		updateSingleState(disp, rootWindow, win, _NET_WM_STATE_MODAL, addRemove)
	}
	if state&STICKY > 0 {
		updateSingleState(disp, rootWindow, win, _NET_WM_STATE_STICKY, addRemove)
	}
	if state&MAXIMIZED_VERT > 0 {
		updateSingleState(disp, rootWindow, win, _NET_WM_STATE_MAXIMIZED_VERT, addRemove)
	}
	if state&MAXIMIZED_HORZ > 0 {
		updateSingleState(disp, rootWindow, win, _NET_WM_STATE_MAXIMIZED_HORZ, addRemove)
	}
	if state&SHADED > 0 {
		updateSingleState(disp, rootWindow, win, _NET_WM_STATE_SHADED, addRemove)
	}
	if state&SKIP_TASKBAR > 0 {
		updateSingleState(disp, rootWindow, win, _NET_WM_STATE_SKIP_TASKBAR, addRemove)
	}
	if state&SKIP_PAGER > 0 {
		updateSingleState(disp, rootWindow, win, _NET_WM_STATE_SKIP_PAGER, addRemove)
	}
	if state&HIDDEN > 0 {
		updateSingleState(disp, rootWindow, win, _NET_WM_STATE_HIDDEN, addRemove)
	}
	if state&FULLSCREEN > 0 {
		updateSingleState(disp, rootWindow, win, _NET_WM_STATE_FULLSCREEN, addRemove)
	}
	if state&ABOVE > 0 {
		updateSingleState(disp, rootWindow, win, _NET_WM_STATE_ABOVE, addRemove)
	}
	if state&BELOW > 0 {
		updateSingleState(disp, rootWindow, win, _NET_WM_STATE_BELOW, addRemove)
	}
	if state&DEMANDS_ATTENTION > 0 {
		updateSingleState(disp, rootWindow, win, _NET_WM_STATE_DEMANDS_ATTENTION, addRemove)
	}
	C.XFlush(disp)

}

func updateSingleState(disp *C.Display, rootWindow C.Window, win C.Window, atom C.Atom, addRemove C.long) {
	var event = C.createClientMessage32(win, _NET_WM_STATE, addRemove, C.long(atom), 0, 0, 0)
	C.XSendEvent(disp, rootWindow, 0, redirectAndNotifyMask, &event)
}

func getGeometry(disp *C.Display, wId uint32) (int32, int32, uint32, uint32, error) {
	var root C.ulong
	var x C.int
	var y C.int
	var width C.uint
	var height C.uint
	var border_width C.uint
	var depth C.uint

	if status := C.XGetGeometry(disp, C.ulong(wId), &root, &x, &y, &width, &height, &border_width, &depth); status != 0 {
		return int32(x), int32(y), uint32(width), uint32(height), nil
	} else {
		return 0, 0, 0, 0, fmt.Errorf("Could not get geometry\n")
	}
}

func getBytes(disp *C.Display, ulong_window C.ulong, prop C.Atom) ([]byte, error) {
	var long_offset C.long
	var long_length = C.long(256)

	var result []byte
	var actual_type_return C.Atom
	var actual_format_return C.int
	var nitems_return C.ulong
	var bytes_after_return C.ulong
	var prop_return *C.uchar
	for {
		var status = C.XGetWindowProperty(disp, ulong_window, prop, long_offset, long_length, 0, C.AnyPropertyType,
			&actual_type_return, &actual_format_return, &nitems_return, &bytes_after_return, &prop_return)

		if err := CheckError(status); err != nil {
			return nil, err
		} else if actual_format_return != 8 {
			return nil, errors.New(fmt.Sprintf("Expected format 8, got %d", actual_format_return))
		}

		var currentLen = len(result)
		var growBy = int(nitems_return)
		var neededCapacity = currentLen + growBy

		if cap(result) < neededCapacity {
			tmp := make([]byte, currentLen, neededCapacity)
			for i := 0; i < currentLen; i++ {
				tmp[i] = result[i]
			}
			result = tmp
		}

		for i := 0; i < growBy; i++ {
			result = append(result, byte(C.getByte(prop_return, C.int(i))))
		}

		C.XFree(unsafe.Pointer(prop_return))

		if bytes_after_return == 0 {
			return result, nil
		}
		long_length = C.long(bytes_after_return)/4 + 1
		long_offset = long_offset + C.long(nitems_return)*4
	}
}

func getUint32s(disp *C.Display, ulong_window C.ulong, prop C.Atom) ([]uint32, error) {
	//var ulong_window = C.ulong(window)

	var long_offset C.long
	var long_length = C.long(256)

	var result []uint32
	var actual_type_return C.Atom
	var actual_format_return C.int
	var nitems_return C.ulong
	var bytes_after_return C.ulong
	var prop_return *C.uchar
	for {
		var error = C.XGetWindowProperty(disp, ulong_window, prop, long_offset, long_length, 0, C.AnyPropertyType,
			&actual_type_return, &actual_format_return, &nitems_return, &bytes_after_return, &prop_return)

		if err := CheckError(error); err != nil {
			return nil, err
		} else if actual_format_return != 32 {
			return nil, errors.New(fmt.Sprintf("Expected format 32, got %d", actual_format_return))
		}

		var currentLen = len(result)
		var growBy = int(nitems_return)
		var neededCapacity = currentLen + growBy

		if cap(result) < neededCapacity {
			tmp := make([]uint32, currentLen, neededCapacity)
			for i := 0; i < currentLen; i++ {
				tmp[i] = result[i]
			}
			result = tmp
		}

		for i := 0; i < growBy; i++ {
			result = append(result, uint32(C.getLong(prop_return, C.int(i))))
		}

		C.XFree(unsafe.Pointer(prop_return))

		if bytes_after_return == 0 {
			return result, nil
		}

		long_length = C.long(bytes_after_return)/4 + 1
		long_offset = long_offset + C.long(nitems_return)
	}
}

func getScreenShotAsXImage(disp *C.Display, wId uint32) (*C.XImage, uint32, uint32) {
	var _, _, w, h, err = getGeometry(disp, wId)
	if err != nil {
		return nil, 0, 0
	}
	return C.XGetImage(disp, C.ulong(wId), C.int(0), C.int(0), C.uint(w), C.uint(h), C.AllPlanes, C.ZPixmap), w, h
}

func internAtom(disp *C.Display, name string) C.ulong {
	var cName = C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return C.XInternAtom(disp, cName, 1)
}

func getAtomName(disp *C.Display, atom C.ulong) string {
	var tmp = C.XGetAtomName(disp, atom)
	defer C.XFree(unsafe.Pointer(tmp))
	return C.GoString(tmp)
}

// ------------------------------------------------------------------------------------------------------------------

// Some commonly used atoms
var _NET_ACTIVE_WINDOW = internAtom(commonProxy.disp, "_NET_ACTIVE_WINDOW")
var _NET_CLOSE_WINDOW = internAtom(commonProxy.disp, "_NET_CLOSE_WINDOW")
var _NET_WM_VISIBLE_NAME = internAtom(commonProxy.disp, "_NET_WM_VISIBLE_NAME")
var _NET_WM_NAME = internAtom(commonProxy.disp, "_NET_WM_NAME")
var _NET_WM_PID = internAtom(commonProxy.disp, "_NET_WM_PID")
var _WM_NAME = internAtom(commonProxy.disp, "_WM_NAME")
var _NET_WM_VISIBLE_ICON_NAME = internAtom(commonProxy.disp, "_NET_WM_VISIBLE_ICON_NAME")
var _NET_WM_ICON_NAME = internAtom(commonProxy.disp, "_NET_WM_ICON_NAME")
var _NET_WM_ICON = internAtom(commonProxy.disp, "_NET_WM_ICON")
var _NET_CLIENT_LIST_STACKING = internAtom(commonProxy.disp, "_NET_CLIENT_LIST_STACKING")
var _NET_DESKTOP_GEOMETRY = internAtom(commonProxy.disp, "_NET_DESKTOP_GEOMETRY")
var _NET_WM_STATE = internAtom(commonProxy.disp, "_NET_WM_STATE")
var _NET_WM_STATE_MODAL = internAtom(commonProxy.disp, "_NET_WM_STATE_MODAL")
var _NET_WM_STATE_STICKY = internAtom(commonProxy.disp, "_NET_WM_STATE_STICKY")
var _NET_WM_STATE_MAXIMIZED_VERT = internAtom(commonProxy.disp, "_NET_WM_STATE_MAXIMIZED_VERT")
var _NET_WM_STATE_MAXIMIZED_HORZ = internAtom(commonProxy.disp, "_NET_WM_STATE_MAXIMIZED_HORZ")
var _NET_WM_STATE_SHADED = internAtom(commonProxy.disp, "_NET_WM_STATE_SHADED")
var _NET_WM_STATE_SKIP_TASKBAR = internAtom(commonProxy.disp, "_NET_WM_STATE_SKIP_TASKBAR")
var _NET_WM_STATE_SKIP_PAGER = internAtom(commonProxy.disp, "_NET_WM_STATE_SKIP_PAGER")
var _NET_WM_STATE_HIDDEN = internAtom(commonProxy.disp, "_NET_WM_STATE_HIDDEN")
var _NET_WM_STATE_FULLSCREEN = internAtom(commonProxy.disp, "_NET_WM_STATE_FULLSCREEN")
var _NET_WM_STATE_ABOVE = internAtom(commonProxy.disp, "_NET_WM_STATE_ABOVE")
var _NET_WM_STATE_BELOW = internAtom(commonProxy.disp, "_NET_WM_STATE_BELOW")
var _NET_WM_STATE_DEMANDS_ATTENTION = internAtom(commonProxy.disp, "_NET_WM_STATE_DEMANDS_ATTENTION")
var _NET_WM_WINDOW_OPACITY = internAtom(commonProxy.disp, "_NET_WM_WINDOW_OPACITY")
var _NET_RESTACK_WINDOW = internAtom(commonProxy.disp, "_NET_RESTACK_WINDOW")
var XA_CARDINAL = internAtom(commonProxy.disp, "XA_CARDINAL")

func init() {
	C.XSetErrorHandler(C.XErrorHandler(C.forgiving_X_error_handler))
	var disp = C.XOpenDisplay(nil)
	_NET_ACTIVE_WINDOW = internAtom(disp, "_NET_ACTIVE_WINDOW")
	_NET_CLOSE_WINDOW = internAtom(disp, "_NET_CLOSE_WINDOW")
	_NET_WM_VISIBLE_NAME = internAtom(disp, "_NET_WM_VISIBLE_NAME")
	_NET_WM_NAME = internAtom(disp, "_NET_WM_NAME")
	_NET_WM_PID = internAtom(disp, "_NET_WM_PID")
	_WM_NAME = internAtom(disp, "_WM_NAME")
	_NET_WM_VISIBLE_ICON_NAME = internAtom(disp, "_NET_WM_VISIBLE_ICON_NAME")
	_NET_WM_ICON_NAME = internAtom(disp, "_NET_WM_ICON_NAME")
	_NET_WM_ICON = internAtom(disp, "_NET_WM_ICON")
	_NET_CLIENT_LIST_STACKING = internAtom(disp, "_NET_CLIENT_LIST_STACKING")
	_NET_DESKTOP_GEOMETRY = internAtom(disp, "_NET_DESKTOP_GEOMETRY")
	_NET_WM_STATE = internAtom(disp, "_NET_WM_STATE")
	_NET_WM_STATE_MODAL = internAtom(disp, "_NET_WM_STATE_MODAL")
	_NET_WM_STATE_STICKY = internAtom(disp, "_NET_WM_STATE_STICKY")
	_NET_WM_STATE_MAXIMIZED_VERT = internAtom(disp, "_NET_WM_STATE_MAXIMIZED_VERT")
	_NET_WM_STATE_MAXIMIZED_HORZ = internAtom(disp, "_NET_WM_STATE_MAXIMIZED_HORZ")
	_NET_WM_STATE_SHADED = internAtom(disp, "_NET_WM_STATE_SHADED")
	_NET_WM_STATE_SKIP_TASKBAR = internAtom(disp, "_NET_WM_STATE_SKIP_TASKBAR")
	_NET_WM_STATE_SKIP_PAGER = internAtom(disp, "_NET_WM_STATE_SKIP_PAGER")
	_NET_WM_STATE_HIDDEN = internAtom(disp, "_NET_WM_STATE_HIDDEN")
	_NET_WM_STATE_FULLSCREEN = internAtom(disp, "_NET_WM_STATE_FULLSCREEN")
	_NET_WM_STATE_ABOVE = internAtom(disp, "_NET_WM_STATE_ABOVE")
	_NET_WM_STATE_BELOW = internAtom(disp, "_NET_WM_STATE_BELOW")
	_NET_WM_STATE_DEMANDS_ATTENTION = internAtom(disp, "_NET_WM_STATE_DEMANDS_ATTENTION")
	_NET_WM_WINDOW_OPACITY = internAtom(disp, "_NET_WM_WINDOW_OPACITY")
	_NET_RESTACK_WINDOW = internAtom(disp, "_NET_RESTACK_WINDOW")
	XA_CARDINAL = internAtom(disp, "XA_CARDINAL")
	C.XFree(unsafe.Pointer(disp))
}

func CheckError(error C.int) error {
	switch error {
	case 0:
		return nil
	case C.BadAlloc:
		return errors.New("The server failed to allocate the requested resource or server memory.")
	case C.BadAtom:
		return errors.New("A value for an Atom argument does not name a defined Atom.")
	case C.BadValue:
		return errors.New("Some numeric value falls outside the range of values accepted by the request. Unless a specific range is specified for an argument, the full range defined by the argument's type is accepted. Any argument defined as a set of alternatives can generate this error.")
	case C.BadWindow:
		return errors.New(fmt.Sprintf("A value for a Window argument does not name a defined Window"))
	default:
		return errors.New(fmt.Sprintf("Uknown error: %d", error))
	}
}
