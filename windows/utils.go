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
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"unsafe"
)

func init() {
	C.XSetErrorHandler(C.XErrorHandler(C.forgiving_X_error_handler))
}

const (
	NET_WM_VISIBLE_NAME         = "_NET_WM_VISIBLE_NAME"
	NET_WM_NAME                 = "_NET_WM_NAME"
	WM_NAME                     = "WM_NAME"
	NET_WM_ICON                 = "_NET_WM_ICON"
	NET_CLIENT_LIST_STACKING    = "_NET_CLIENT_LIST_STACKING"
	NET_WM_STATE                = "_NET_WM_STATE"
	NET_DESKTOP_GEOMETRY        = "_NET_DESKTOP_GEOMETRY"
	NET_WM_STATE_MAXIMIZED_VERT = "_NET_WM_STATE_MAXIMIZED_VERT"
	NET_WM_STATE_MAXIMIZED_HORZ = "_NET_WM_STATE_MAXIMIZED_HORZ"
	NET_WM_STATE_HIDDEN         = "_NET_WM_STATE_HIDDEN"
)

// Either 'Property' or X,Y,W,H will be set
type Event struct {
	Window     uint32
	Property   string
	X, Y, W, H int
}

/**
 * Wrapper around connections to X11. Not threadsafe, caller must take lock
 */
type Connection struct {
	display    *C.Display
	rootWindow C.Window

	atomCache     map[string]C.Atom
	atomNameCache map[C.Atom]string
}

func MakeConnection() *Connection {
	var conn = Connection{}
	conn.display = C.XOpenDisplay(nil)
	var defaultScreen = C.ds(conn.display)
	conn.rootWindow = C.rw(conn.display, defaultScreen)
	conn.atomCache = make(map[string]C.Atom)
	conn.atomNameCache = make(map[C.Atom]string)
	return &conn
}

func (c *Connection) SubscribeToEvents() {
	C.XSelectInput(c.display, c.rootWindow, C.PropertyChangeMask)
	C.XRRSelectInput(c.display, c.rootWindow, C.RRScreenChangeNotifyMask)
}

// Will hang until an event arrives
// Returns
//   0: error
//   1: monitor/screen change
//   2: window stack change
func WaitForEvent(c *Connection) int {
	var event C.XEvent
	for {
		if err := CheckError(C.XNextEvent(c.display, &event)); err != nil {
			fmt.Println("WARN error getting event from X:", err)
			return 0
		} else if C.getType(&event) != C.PropertyNotify {
			continue
		} else {
			var name = c.atomName(C.xproperty(&event).atom)
			switch name {
			case NET_DESKTOP_GEOMETRY:
				return 1
			case NET_CLIENT_LIST_STACKING:
				return 2
			}
		}
	}
}

// Will hang until either a property change or a configure event happens
func (c *Connection) NextEvent() (Event, error) {
	var event C.XEvent
	for {
		if err := CheckError(C.XNextEvent(c.display, &event)); err != nil {
			return Event{}, err // TODO Maybe some max error count and then a panic.. To prevent looping at 100% cpu
		} else {
			switch C.getType(&event) {
			case C.PropertyNotify:
				var xproperty = C.xproperty(&event)
				return Event{Window: uint32(xproperty.window), Property: c.atomName(xproperty.atom)}, nil
			case C.ConfigureNotify:
				var xconfigure = C.xconfigure(&event)
				return Event{Window: uint32(xconfigure.window),
					X: int(xconfigure.x), Y: int(xconfigure.y), W: int(xconfigure.width), H: int(xconfigure.height)}, nil
			}
		}
	}
}

func (c *Connection) atom(name string) C.Atom {
	if val, ok := c.atomCache[name]; ok {
		return val
	} else {
		var cName = C.CString(name)
		defer C.free(unsafe.Pointer(cName))
		val = C.XInternAtom(c.display, cName, 1)
		if val == C.None {
			log.Fatal(fmt.Sprintf("Atom %s does not exist", name))
		}
		c.atomCache[name] = val
		return val
	}

}

func (c *Connection) atomName(atom C.Atom) string {
	if name, ok := c.atomNameCache[atom]; ok {
		return name
	} else {
		var tmp = C.XGetAtomName(c.display, atom)
		defer C.XFree(unsafe.Pointer(tmp))
		c.atomNameCache[atom] = C.GoString(tmp)
		return c.atomNameCache[atom]
	}
}

func (c *Connection) GetBytes(window uint32, property string) ([]byte, error) {
	var ulong_window = C.ulong(window)
	if ulong_window == 0 {
		ulong_window = c.rootWindow
	}
	var prop = c.atom(property)
	var long_offset C.long
	var long_length = C.long(256)

	var result []byte
	var actual_type_return C.Atom
	var actual_format_return C.int
	var nitems_return C.ulong
	var bytes_after_return C.ulong
	var prop_return *C.uchar
	for {
		var status = C.XGetWindowProperty(c.display, ulong_window, prop, long_offset, long_length, 0, C.AnyPropertyType,
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

func (c *Connection) GetPropStr(wId uint32, property string) (string, error) {
	bytes, err := c.GetBytes(wId, property)
	return string(bytes), err
}

func (c *Connection) GetUint32s(window uint32, property string) ([]uint32, error) {
	var ulong_window = C.ulong(window)
	if ulong_window == 0 {
		ulong_window = c.rootWindow
	}
	var prop = c.atom(property)
	var long_offset C.long
	var long_length = C.long(256)

	var result []uint32
	var actual_type_return C.Atom
	var actual_format_return C.int
	var nitems_return C.ulong
	var bytes_after_return C.ulong
	var prop_return *C.uchar
	for {
		var error = C.XGetWindowProperty(c.display, ulong_window, prop, long_offset, long_length, 0, C.AnyPropertyType,
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

func (c *Connection) GetAtoms(wId uint32, property string) ([]string, error) {
	if atoms, err := c.GetUint32s(wId, property); err != nil {
		return nil, err
	} else {
		var states = make([]string, len(atoms), len(atoms))
		for i, atom := range atoms {
			states[i] = c.atomName(C.ulong(atom))
		}
		return states, nil
	}
}

func (c *Connection) GetParent(wId uint32) (uint32, error) {
	var root_return C.ulong
	var parent_return C.ulong
	var children_return *C.ulong
	var nchildren_return C.uint
	for {
		if C.XQueryTree(c.display, C.ulong(wId), &root_return, &parent_return, &children_return, &nchildren_return) == 0 {
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

	var status = C.XGetGeometry(c.display, C.ulong(wId), &root, &x, &y, &width, &height, &border_width, &depth)
	if status != 0 {
		return int32(x), int32(y), uint32(width), uint32(height), nil
	} else {
		return 0, 0, 0, 0, fmt.Errorf("Could not get geometry\n")
	}
}

// ---------------------------------------------------------------------------------------------

func GetStack(c *Connection) []uint32 {
	if tmp, err := c.GetUint32s(0, NET_CLIENT_LIST_STACKING); err != nil {
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

func (c *Connection) GetName(wId uint32) (string, error) {
	if bytes, err := c.GetBytes(wId, NET_WM_VISIBLE_NAME); err == nil {
		return string(bytes), nil
	} else if bytes, err = c.GetBytes(wId, NET_WM_NAME); err == nil {
		return string(bytes), nil
	} else if bytes, err = c.GetBytes(wId, WM_NAME); err == nil {
		return string(bytes), nil
	} else {
		return "", errors.New("Neither '_NET_WM_VISIBLE_NAME', '_NET_WM_NAME' nor 'WM_NAME' set")
	}
}

func (c *Connection) GetIcon(wId uint32) ([]uint32, error) {
	return c.GetUint32s(wId, NET_WM_ICON)
}

func (c *Connection) GetState(wId uint32) ([]string, error) {
	return c.GetAtoms(wId, NET_WM_STATE)
}

func (c *Connection) SendEvent(wId uint32, atom string, l0 int32, l1 int32, l2 int32, l3 int32, l4 int32) {
	var event = C.createClientMessage32(C.Window(wId), c.atom(atom), C.long(l0), C.long(l1), C.long(l2), C.long(l3), C.long(l4))
	var mask C.long = C.SubstructureRedirectMask | C.SubstructureNotifyMask
	C.XSendEvent(c.display, c.rootWindow, 0, mask, &event)
	C.XFlush(c.display)
}

func (c *Connection) AddWmState(wId uint32, atom string) {
	var event = C.createClientMessage32(C.Window(wId), c.atom("_NET_WM_STATE"), 2, C.long(c.atom(atom)), 0, 0, 0)
	var mask C.long = C.SubstructureRedirectMask | C.SubstructureNotifyMask
	C.XSendEvent(c.display, c.rootWindow, 0, mask, &event)
	C.XFlush(c.display)
}

func (c *Connection) RemoveWmState(wId uint32, atom string) {
	var event = C.createClientMessage32(C.Window(wId), c.atom("_NET_WM_STATE"), 2, C.long(c.atom(atom)), 0, 0, 0)
	var mask C.long = C.SubstructureRedirectMask | C.SubstructureNotifyMask
	C.XSendEvent(c.display, c.rootWindow, 0, mask, &event)
	C.XFlush(c.display)
}

func (c *Connection) MoveToMonitor(wId uint32, monitorName string) error {
	monitorList := monitors.Load().([]*Monitor)
	for _, m := range monitorList {
		if monitorName == m.Name {
			newX, newY := m.X+int32(m.W)/10, m.Y+int32(m.H)/10
			newW, newH := m.W/10*8, m.H/10*8
			if _, _, winW, winH, err := c.GetGeometry(wId); err == nil {
				if winW < newW {
					newW = winW
				}

				if winH < newH {
					newH = winH
				}
			}
			C.XMoveResizeWindow(c.display, C.Window(wId), C.int(newX), C.int(newY), C.uint(newW), C.uint(newH))
			C.XFlush(c.display)
			return nil
		}
	}
	return fmt.Errorf("No such monitor %s", monitorName)
}

func (c *Connection) Move(wId uint32, x int32, y int32, width int32, height int32) {
	/*	var event = C.createConfigureMessage32(C.Window(wId), C.Window(c.rootWindow), C.int(x), C.int(y), C.int(width), C.int(height))
		var mask C.long = C.SubstructureRedirectMask | C.SubstructureNotifyMask
		C.XSendEvent(c.display, c.rootWindow, 0, mask, &event)
		C.XFlush(c.display)*/
	fmt.Println("move")
	C.XFlush(c.display)
}

func (c *Connection) RaiseAndFocusWindow(wId uint32) {
	var event = C.createClientMessage32(C.Window(wId), c.atom("_NET_ACTIVE_WINDOW"), 2, 0, 0, 0, 0)
	var mask C.long = C.SubstructureRedirectMask | C.SubstructureNotifyMask
	C.XSendEvent(c.display, c.rootWindow, 0, mask, &event)
	C.XFlush(c.display)
}

func (c *Connection) CloseWindow(wId uint32) {
	var event = C.createClientMessage32(C.Window(wId), c.atom("_NET_CLOSE_WINDOW"), 2, 0, 0, 0, 0)
	var mask C.long = C.SubstructureRedirectMask | C.SubstructureNotifyMask
	log.Println("Sending close event")
	C.XSendEvent(c.display, c.rootWindow, 0, mask, &event)
	C.XFlush(c.display)
	log.Println("Close event sent")
}

func (c *Connection) GetScreenshotAsPng(wId uint32, downscale uint8) ([]byte, error) {
	var _, _, w, h, err = c.GetGeometry(wId)
	if err != nil {
		return nil, err
	}

	var ximage = C.XGetImage(c.display, C.ulong(wId), C.int(0), C.int(0), C.uint(w), C.uint(h), C.AllPlanes, C.ZPixmap)
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

func GetMonitors(c *Connection) []*Monitor {
	var num C.int
	xrrmonitorsPtr := C.XRRGetMonitors(c.display, c.rootWindow, 1, &num)
	xrrmonitorsArr := (*[1 << 30]C.XRRMonitorInfo)(unsafe.Pointer(xrrmonitorsPtr))
	var monitors = make([]*Monitor, num, num)
	var bound int = int(num)
	for i := 0; i < bound; i++ {
		monitors[i] = &Monitor{
			X:    int32(xrrmonitorsArr[i].x),
			Y:    int32(xrrmonitorsArr[i].y),
			W:    uint32(xrrmonitorsArr[i].width),
			H:    uint32(xrrmonitorsArr[i].height),
			Wmm:  uint32(xrrmonitorsArr[i].mwidth),
			Hmm:  uint32(xrrmonitorsArr[i].mheight),
			Name: c.atomName(xrrmonitorsArr[i].name),
		}
	}
	C.XRRFreeMonitors(xrrmonitorsPtr)
	return monitors
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
