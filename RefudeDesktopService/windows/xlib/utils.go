// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package xlib

/**
 * All communication with xlib (and X) happens through this package
 */

/*
#cgo LDFLAGS: -lX11
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <X11/Xlib.h>
// Cant use 'type' in Go, hence...
inline int getType(XEvent* e) { return e->type; }

// Using C macros in Go seems tricky, so..
inline int ds(Display* d) { return DefaultScreen(d); }
inline Window rw(Display *d, int screen) { return RootWindow(d, screen); }

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
	"errors"
	"fmt"
	"log"
	"sync"
	"unsafe"
)

func init() {
	C.XSetErrorHandler(C.XErrorHandler(C.forgiving_X_error_handler))
}

var display = C.XOpenDisplay(nil)
var defaultScreen = C.ds(display)
var rootWindow = C.rw(display, defaultScreen)

var atomCache = make(map[string]C.Atom)
var atomNameCache = make(map[C.Atom]string)

// Either 'Property' or X,Y,W,H will be set
type Event struct {
	Window     uint32
	Property   string
	X, Y, W, H int
}

func atom(name string) C.Atom {
	if val, ok := atomCache[name]; ok {
		return val
	} else {
		var cName = C.CString(name)
		defer C.free(unsafe.Pointer(cName))
		val = C.XInternAtom(display, cName, 1)
		if val == C.None {
			log.Fatal(fmt.Sprintf("Atom %s does not exist", name))
		}
		atomCache[name] = val
		return val
	}
}

func atomName(atom C.Atom) string {
	if name, ok := atomNameCache[atom]; ok {
		return name
	} else {
		var tmp = C.XGetAtomName(display, atom)
		defer C.XFree(unsafe.Pointer(tmp))
		atomNameCache[atom] = C.GoString(tmp)
		return atomNameCache[atom]
	}
}

func getBytes(window uint32, property string) ([]byte, error) {

	var ulong_window = C.ulong(window)
	if ulong_window == 0 {
		ulong_window = rootWindow
	}
	var prop = atom(property)
	var long_offset C.long
	var long_length = C.long(256)

	var result []byte
	var actual_type_return C.Atom
	var actual_format_return C.int
	var nitems_return C.ulong
	var bytes_after_return C.ulong
	var prop_return *C.uchar
	for {
		var status = C.XGetWindowProperty(display, ulong_window, prop, long_offset, long_length, 0, C.AnyPropertyType,
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

func getUint32s(window uint32, property string) ([]uint32, error) {
	var ulong_window = C.ulong(window)
	if ulong_window == 0 {
		ulong_window = rootWindow
	}
	var prop = atom(property)
	var long_offset C.long
	var long_length = C.long(256)

	var result []uint32
	var actual_type_return C.Atom
	var actual_format_return C.int
	var nitems_return C.ulong
	var bytes_after_return C.ulong
	var prop_return *C.uchar
	for {
		var error = C.XGetWindowProperty(display, ulong_window, prop, long_offset, long_length, 0, C.AnyPropertyType,
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

func getAtoms(wId uint32, property string) ([]string, error) {
	if atoms, err := getUint32s(wId, property); err != nil {
		return nil, err
	} else {
		var states = make([]string, len(atoms), len(atoms))
		for i, atom := range atoms {
			states[i] = atomName(C.ulong(atom))
		}
		return states, nil
	}
}

// Search up the tree, via parent relation, till we find a window whos parent is the root window.
// Return that
func getParent(wId uint32) (uint32, error) {
	var root_return C.ulong
	var parent_return C.ulong
	var children_return *C.ulong
	var nchildren_return C.uint
	for {
		if C.XQueryTree(display, C.ulong(wId), &root_return, &parent_return, &children_return, &nchildren_return) == 0 {
			return 0, errors.New("Error from XQueryTree")
		} else {
			if children_return != nil {
				C.XFree(unsafe.Pointer(children_return))
			}
			if parent_return == rootWindow {
				return wId, nil
			} else {
				wId = uint32(parent_return)
			}
		}
	}
}

func getGeometry(wId uint32) (int32, int32, uint32, uint32, error) {
	var root C.ulong
	var x C.int
	var y C.int
	var width C.uint
	var height C.uint
	var border_width C.uint
	var depth C.uint

	var status = C.XGetGeometry(display, C.ulong(wId), &root, &x, &y, &width, &height, &border_width, &depth)
	if status != 0 {
		return int32(x), int32(y), uint32(width), uint32(height), nil
	} else {
		return 0, 0, 0, 0, fmt.Errorf("Could not get geometry\n")
	}
}

// ---------------------------------------------------------------------------------------------
const (
	NET_WM_VISIBLE_NAME      = "_NET_WM_VISIBLE_NAME"
	NET_WM_NAME              = "_NET_WM_NAME"
	WM_NAME                  = "WM_NAME"
	NET_WM_ICON              = "_NET_WM_ICON"
	NET_CLIENT_LIST_STACKING = "_NET_CLIENT_LIST_STACKING"
	NET_WM_STATE             = "_NET_WM_STATE"
)

var mutex sync.Mutex

func GetStack() ([]uint32, error) {
	mutex.Lock()
	defer mutex.Unlock()
	return getUint32s(0, NET_CLIENT_LIST_STACKING)
}

func GetParent(wId uint32) (uint32, error) {
	mutex.Lock()
	defer mutex.Unlock()

	return getParent(wId)
}

func GetGeometry(wId uint32) (int32, int32, uint32, uint32, error) {
	mutex.Lock()
	defer mutex.Unlock()
	return getGeometry(wId)
}

func GetName(wId uint32) (string, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if bytes, err := getBytes(wId, NET_WM_VISIBLE_NAME); err == nil {
		return string(bytes), nil
	} else if bytes, err = getBytes(wId, NET_WM_NAME); err == nil {
		return string(bytes), nil
	} else if bytes, err = getBytes(wId, WM_NAME); err == nil {
		return string(bytes), nil
	} else {
		return "", errors.New("Neither '_NET_WM_VISIBLE_NAME', '_NET_WM_NAME' nor 'WM_NAME' set")
	}
}

func GetIcon(wId uint32) ([]uint32, error) {
	mutex.Lock()
	defer mutex.Unlock()

	return getUint32s(wId, NET_WM_ICON)
}

func GetState(wId uint32) ([]string, error) {
	mutex.Lock()
	defer mutex.Unlock()

	return getAtoms(wId, NET_WM_STATE)
}

func RaiseAndFocusWindow(wId uint32) {
	mutex.Lock()
	defer mutex.Unlock()

	var event = C.createClientMessage32(C.Window(wId), atom("_NET_ACTIVE_WINDOW"), 2, 0, 0, 0, 0)
	var mask C.long = C.SubstructureRedirectMask | C.SubstructureNotifyMask
	C.XSendEvent(display, rootWindow, 0, mask, &event)
	C.XFlush(display)
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
