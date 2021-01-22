// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

/**
 * Most interfacing with c happens through this file
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
	"unsafe"
)

func init() {
	C.XSetErrorHandler(C.XErrorHandler(C.forgiving_X_error_handler))
}

// Some atoms, frequently used
var _NET_ACTIVE_WINDOW,
	_NET_CLOSE_WINDOW,
	_NET_WM_VISIBLE_NAME,
	_NET_WM_NAME,
	_WM_NAME,
	_NET_WM_ICON_NAME,
	_NET_WM_VISIBLE_ICON_NAME,
	_NET_WM_ICON,
	_NET_CLIENT_LIST_STACKING,
	_NET_DESKTOP_GEOMETRY,
	_NET_WM_STATE,
	_NET_WM_STATE_MODAL,
	_NET_WM_STATE_STICKY,
	_NET_WM_STATE_MAXIMIZED_VERT,
	_NET_WM_STATE_MAXIMIZED_HORZ,
	_NET_WM_STATE_SHADED,
	_NET_WM_STATE_SKIP_TASKBAR,
	_NET_WM_STATE_SKIP_PAGER,
	_NET_WM_STATE_HIDDEN,
	_NET_WM_STATE_FULLSCREEN,
	_NET_WM_STATE_ABOVE,
	_NET_WM_STATE_BELOW,
	_NET_WM_STATE_DEMANDS_ATTENTION C.Atom

func init() {
	var c = MakeDisplay()
	_NET_ACTIVE_WINDOW = c.InternAtom("_NET_ACTIVE_WINDOW")
	_NET_CLOSE_WINDOW = c.InternAtom("_NET_CLOSE_WINDOW")
	_NET_WM_VISIBLE_NAME = c.InternAtom("_NET_WM_VISIBLE_NAME")
	_NET_WM_NAME = c.InternAtom("_NET_WM_NAME")
	_WM_NAME = c.InternAtom("_WM_NAME")
	_NET_WM_VISIBLE_ICON_NAME = c.InternAtom("_NET_WM_VISIBLE_ICON_NAME")
	_NET_WM_ICON_NAME = c.InternAtom("_NET_WM_ICON_NAME")
	_NET_WM_ICON = c.InternAtom("_NET_WM_ICON")
	_NET_CLIENT_LIST_STACKING = c.InternAtom("_NET_CLIENT_LIST_STACKING")
	_NET_DESKTOP_GEOMETRY = c.InternAtom("_NET_DESKTOP_GEOMETRY")
	_NET_WM_STATE = c.InternAtom("_NET_WM_STATE")
	_NET_WM_STATE_MODAL = c.InternAtom("_NET_WM_STATE_MODAL")
	_NET_WM_STATE_STICKY = c.InternAtom("_NET_WM_STATE_STICKY")
	_NET_WM_STATE_MAXIMIZED_VERT = c.InternAtom("_NET_WM_STATE_MAXIMIZED_VERT")
	_NET_WM_STATE_MAXIMIZED_HORZ = c.InternAtom("_NET_WM_STATE_MAXIMIZED_HORZ")
	_NET_WM_STATE_SHADED = c.InternAtom("_NET_WM_STATE_SHADED")
	_NET_WM_STATE_SKIP_TASKBAR = c.InternAtom("_NET_WM_STATE_SKIP_TASKBAR")
	_NET_WM_STATE_SKIP_PAGER = c.InternAtom("_NET_WM_STATE_SKIP_PAGER")
	_NET_WM_STATE_HIDDEN = c.InternAtom("_NET_WM_STATE_HIDDEN")
	_NET_WM_STATE_FULLSCREEN = c.InternAtom("_NET_WM_STATE_FULLSCREEN")
	_NET_WM_STATE_ABOVE = c.InternAtom("_NET_WM_STATE_ABOVE")
	_NET_WM_STATE_BELOW = c.InternAtom("_NET_WM_STATE_BELOW")
	_NET_WM_STATE_DEMANDS_ATTENTION = c.InternAtom("_NET_WM_STATE_DEMANDS_ATTENTION")
}

// Either 'Property' or X,Y,W,H will be set
type Event struct {
	Window     uint32
	Property   string
	X, Y, W, H int
}

type Connection struct {
	sync.Mutex
	disp       *C.Display
	rootWindow C.Window
}

func (d *Connection) SelectInput(window C.Window, mask C.long) {
	d.Lock()
	defer d.Unlock()
	C.XSelectInput(d.disp, window, mask)
}

func (d *Connection) RRSelectInput(window C.Window, mask C.int) {
	d.Lock()
	defer d.Unlock()
	C.XRRSelectInput(d.disp, window, mask)
}

func (c *Connection) SendEvent(ev *C.XEvent) {
	c.Lock()
	defer c.Unlock()
	var mask C.long = C.SubstructureRedirectMask | C.SubstructureNotifyMask
	C.XSendEvent(c.disp, c.rootWindow, 0, mask, ev)
	C.XFlush(c.disp)
}

func (d *Connection) NextEvent(event *C.XEvent) C.int {
	d.Lock()
	defer d.Unlock()
	return C.XNextEvent(d.disp, event)
}

func (d *Connection) InternAtom(name string) C.ulong {
	d.Lock()
	defer d.Unlock()
	var cName = C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return C.XInternAtom(d.disp, cName, 1)
}

func (d *Connection) GetAtomName(atom C.ulong) string {
	d.Lock()
	defer d.Unlock()
	var tmp = C.XGetAtomName(d.disp, atom)
	defer C.XFree(unsafe.Pointer(tmp))
	return C.GoString(tmp)
}

func (c *Connection) QueryTree(w C.Window, root_return *C.ulong, parent_return *C.ulong, children_return **C.ulong, nchildren_return *C.uint) C.int {
	c.Lock()
	defer c.Unlock()
	return C.XQueryTree(c.disp, w, root_return, parent_return, children_return, nchildren_return)
}

func (c *Connection) GetBytes(window uint32, prop C.Atom) ([]byte, error) {
	var ulong_window = C.ulong(window)
	if ulong_window == 0 {
		ulong_window = c.rootWindow
	}
	var long_offset C.long
	var long_length = C.long(256)

	var result []byte
	var actual_type_return C.Atom
	var actual_format_return C.int
	var nitems_return C.ulong
	var bytes_after_return C.ulong
	var prop_return *C.uchar
	for {
		c.Lock()
		var status = C.XGetWindowProperty(c.disp, ulong_window, prop, long_offset, long_length, 0, C.AnyPropertyType,
			&actual_type_return, &actual_format_return, &nitems_return, &bytes_after_return, &prop_return)
		c.Unlock()

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

func (c *Connection) GetUint32s(window uint32, prop C.Atom) ([]uint32, error) {
	var ulong_window = C.ulong(window)
	if ulong_window == 0 {
		ulong_window = c.rootWindow
	}
	var long_offset C.long
	var long_length = C.long(256)

	var result []uint32
	var actual_type_return C.Atom
	var actual_format_return C.int
	var nitems_return C.ulong
	var bytes_after_return C.ulong
	var prop_return *C.uchar
	for {
		c.Lock()
		var error = C.XGetWindowProperty(c.disp, ulong_window, prop, long_offset, long_length, 0, C.AnyPropertyType,
			&actual_type_return, &actual_format_return, &nitems_return, &bytes_after_return, &prop_return)
		c.Unlock()

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

type MonitorData struct {
	X, Y     int32
	W, H     uint32
	Wmm, Hmm uint32
	Name     string
}

func (c *Connection) GetMonitorDataList() []*MonitorData {
	var num C.int

	xrrmonitorsPtr := C.XRRGetMonitors(c.disp, c.rootWindow, 1, &num)
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
			Name: c.GetAtomName(xrrmonitorsArr[i].name),
		}
	}
	C.XRRFreeMonitors(xrrmonitorsPtr)
	return monitors
}

// ----------------------------------------------------------------------------------

// Public

func MakeDisplay() *Connection {
	var disp = C.XOpenDisplay(nil)
	var defaultScreen = C.ds(disp)
	var rootWindow = C.rw(disp, defaultScreen)
	return &Connection{disp: disp, rootWindow: rootWindow}
}

func SubscribeToEvents(c *Connection) {
	c.SelectInput(c.rootWindow, C.PropertyChangeMask)
	c.RRSelectInput(c.rootWindow, C.RRScreenChangeNotifyMask)
}

func SubscribeToWindowEvents(c *Connection, wId uint32) {
	c.SelectInput(C.Window(wId), C.PropertyChangeMask)
}

// Will hang until either a property change or a configure event happens
type EventType uint8

const (
	Unknown EventType = iota
	DesktopGeometry
	DesktopStacking
	WindowTitle
	WindowIconName
	WindowGeometry
	WindowSt
)

func NextEvent(c *Connection) (EventType, uint32, error) {
	var event C.XEvent
	for {
		if err := CheckError(c.NextEvent(&event)); err != nil {
			return Unknown, 0, err
		} else if C.getType(&event) == C.PropertyNotify {
			var xproperty = C.xproperty(&event)
			if xproperty.atom == _NET_CLIENT_LIST_STACKING {
				return DesktopStacking, 0, nil
			} else if xproperty.atom == _NET_DESKTOP_GEOMETRY {
				return DesktopGeometry, 0, nil
			} else if xproperty.atom == _NET_WM_VISIBLE_NAME || xproperty.atom == _NET_WM_NAME || xproperty.atom == _WM_NAME {
				return WindowTitle, uint32(xproperty.window), nil
			} else if xproperty.atom == _NET_WM_VISIBLE_ICON_NAME || xproperty.atom == _NET_WM_ICON_NAME || xproperty.atom == _NET_WM_ICON {
				return WindowIconName, uint32(xproperty.window), nil
			} else if xproperty.atom == _NET_WM_STATE {
				return WindowSt, uint32(xproperty.window), nil
			}
		} else if C.getType(&event) == C.ConfigureNotify {
			var xconfigure = C.xconfigure(&event)
			return WindowGeometry, uint32(xconfigure.window), nil
		}
	}
}

/*func (c *Display) GetAtoms(wId uint32, property string) ([]string, error) {
	if atoms, err := c.GetUint32s(wId, property); err != nil {
		return nil, err
	} else {
		var states = make([]string, len(atoms), len(atoms))
		for i, atom := range atoms {
			states[i] = c.atomName(C.ulong(atom))
		}
		return states, nil
	}
}*/

func GetParent(c *Connection, wId uint32) (uint32, error) {
	var root_return C.ulong
	var parent_return C.ulong
	var children_return *C.ulong
	var nchildren_return C.uint
	for {
		if c.QueryTree(C.ulong(wId), &root_return, &parent_return, &children_return, &nchildren_return) == 0 {
			return 0, errors.New("Error from XQueryTree")
		} else {
			if children_return != nil {
				C.XFree(unsafe.Pointer(children_return))
			}
			if parent_return == c.rootWindow {
				return wId, nil
			} else {
				wId = uint32(parent_return)
			}
		}
	}
}

func (c *Connection) GetGeometry(wId uint32) (int32, int32, uint32, uint32, error) {
	var root C.ulong
	var x C.int
	var y C.int
	var width C.uint
	var height C.uint
	var border_width C.uint
	var depth C.uint

	c.Lock()
	var status = C.XGetGeometry(c.disp, C.ulong(wId), &root, &x, &y, &width, &height, &border_width, &depth)
	c.Unlock()

	if status != 0 {
		return int32(x), int32(y), uint32(width), uint32(height), nil
	} else {
		return 0, 0, 0, 0, fmt.Errorf("Could not get geometry\n")
	}
}

func (c *Connection) MoveResizeWindow(win C.Window, x C.int, y C.int, w C.uint, h C.uint) {
	c.Lock()
	defer c.Unlock()
	C.XMoveResizeWindow(c.disp, win, x, y, w, h)
	C.XFlush(c.disp)
}

func (c *Connection) UpdateSingleState(wId uint32, atom C.Atom, addRemove C.int) {
	var event = C.createClientMessage32(C.Window(wId), _NET_WM_STATE, 2, C.long(atom), 0, 0, 0)
	c.SendEvent(&event)
}

func (c *Connection) GetImage(wId C.Window, w C.uint, h C.uint) *C.XImage {
	c.Lock()
	defer c.Unlock()
	return C.XGetImage(c.disp, wId, C.int(0), C.int(0), w, h, C.AllPlanes, C.ZPixmap)
}

// ---------------------------------------------------------------------------------------------

func GetStack(c *Connection) []uint32 {
	if tmp, err := c.GetUint32s(0, _NET_CLIENT_LIST_STACKING); err != nil {
		fmt.Println("Error getting stack:", err)
		return []uint32{}
	} else {
		for i := 0; i < len(tmp)/2; i++ {
			j := len(tmp) - i - 1
			tmp[i], tmp[j] = tmp[j], tmp[i]
		}
		return tmp
	}

}

func GetName(c *Connection, wId uint32) (string, error) {
	if bytes, err := c.GetBytes(wId, _NET_WM_VISIBLE_NAME); err == nil {
		return string(bytes), nil
	} else if bytes, err = c.GetBytes(wId, _NET_WM_NAME); err == nil {
		return string(bytes), nil
	} else if bytes, err = c.GetBytes(wId, _WM_NAME); err == nil {
		return string(bytes), nil
	} else {
		return "", errors.New("Neither '_NET_WM_VISIBLE_NAME', '_NET_WM_NAME' nor 'WM_NAME' set")
	}
}

func GetIcon(c *Connection, wId uint32) ([]uint32, error) {
	return c.GetUint32s(wId, _NET_WM_ICON)
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

func GetState(c *Connection, wId uint32) (WindowStateMask, error) {
	var state WindowStateMask = 0
	if atoms, err := c.GetUint32s(wId, _NET_WM_STATE); err != nil {
		return 0, err
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
		return state, nil
	}
}

func AddStates(c *Connection, wId uint32, states WindowStateMask) {
	UpdateState(c, wId, states, 1)
}

func RemoveStates(c *Connection, wId uint32, states WindowStateMask) {
	UpdateState(c, wId, states, 0)
}

func UpdateState(c *Connection, wId uint32, state WindowStateMask, addRemove C.int) {
	if state&MODAL > 0 {
		c.UpdateSingleState(wId, _NET_WM_STATE_MODAL, addRemove)
	}
	if state&STICKY > 0 {
		c.UpdateSingleState(wId, _NET_WM_STATE_STICKY, addRemove)
	}
	if state&MAXIMIZED_VERT > 0 {
		c.UpdateSingleState(wId, _NET_WM_STATE_MAXIMIZED_VERT, addRemove)
	}
	if state&MAXIMIZED_HORZ > 0 {
		c.UpdateSingleState(wId, _NET_WM_STATE_MAXIMIZED_HORZ, addRemove)
	}
	if state&SHADED > 0 {
		c.UpdateSingleState(wId, _NET_WM_STATE_SHADED, addRemove)
	}
	if state&SKIP_TASKBAR > 0 {
		c.UpdateSingleState(wId, _NET_WM_STATE_SKIP_TASKBAR, addRemove)
	}
	if state&SKIP_PAGER > 0 {
		c.UpdateSingleState(wId, _NET_WM_STATE_SKIP_PAGER, addRemove)
	}
	if state&HIDDEN > 0 {
		c.UpdateSingleState(wId, _NET_WM_STATE_HIDDEN, addRemove)
	}
	if state&FULLSCREEN > 0 {
		c.UpdateSingleState(wId, _NET_WM_STATE_FULLSCREEN, addRemove)
	}
	if state&ABOVE > 0 {
		c.UpdateSingleState(wId, _NET_WM_STATE_ABOVE, addRemove)
	}
	if state&BELOW > 0 {
		c.UpdateSingleState(wId, _NET_WM_STATE_BELOW, addRemove)
	}
	if state&DEMANDS_ATTENTION > 0 {
		c.UpdateSingleState(wId, _NET_WM_STATE_DEMANDS_ATTENTION, addRemove)
	}

}

func SetBounds(c *Connection, wId uint32, x int32, y int32, w uint32, h uint32) {
	c.MoveResizeWindow(C.Window(wId), C.int(x), C.int(y), C.uint(w), C.uint(h))
}

func RaiseAndFocusWindow(c *Connection, wId uint32) {
	var event = C.createClientMessage32(C.Window(wId), _NET_ACTIVE_WINDOW, 2, 0, 0, 0, 0)
	c.SendEvent(&event)
}

func (c *Connection) CloseWindow(wId uint32) {
	var event = C.createClientMessage32(C.Window(wId), _NET_CLOSE_WINDOW, 2, 0, 0, 0, 0)
	c.SendEvent(&event)
}

func GetScreenshotAsPng(c *Connection, wId uint32, downscale uint8) ([]byte, error) {
	var _, _, w, h, err = c.GetGeometry(wId)
	if err != nil {
		return nil, err
	}

	var ximage = c.GetImage(C.ulong(wId), C.uint(w), C.uint(h))
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

// ------------------------------------------------------------------------------------------------------------------

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
