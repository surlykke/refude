// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package wayland

// gtk4-layer-shell is not used here, but in notifygui.c. It needs to be loaded
// before wayland-client, hence its added to LDFLAGS

/*
#cgo CFLAGS: -I. -I${SRCDIR}
#cgo LDFLAGS: -lgtk4-layer-shell -lwayland-client

#include <stdint.h>
#include <wayland-client-core.h>
#include "wlr-foreign-toplevel-management-unstable-v1-client-protocol.h"

extern struct wl_display *wl_display;

void initManager();
int wl_display_dispatch();
typedef struct wl_output* wl_output;
typedef struct zwlr_foreign_toplevel_handle_v1 *toplevel_handle;
typedef struct wl_array* wl_array;
void close_toplevel(uintptr_t);
void activate_toplevel(uintptr_t);
void hide_toplevel(uintptr_t);
void show_toplevel(uintptr_t);
*/
import "C"
import (
	"unsafe"
)

func close(handle uint64) {
	C.close_toplevel(C.uintptr_t(handle))
}

func activate(handle uint64) {
	C.activate_toplevel(C.uintptr_t(handle))
}

func hide(handle uint64) {
	C.hide_toplevel(C.uintptr_t(handle))
}

func show(handle uint64) {
	C.show_toplevel(C.uintptr_t(handle))
}

//export handle_title
func handle_title(handle C.uintptr_t, c_title *C.char) {
	windowUpdates <- windowUpdate{wId: uint64(handle), title: C.GoString(c_title)}
}

//export handle_app_id
func handle_app_id(handle C.uintptr_t, c_app_id *C.char) {
	windowUpdates <- windowUpdate{wId: uint64(handle), appId: C.GoString(c_app_id)}
}

//export handle_output_enter
func handle_output_enter(handle C.uintptr_t, output C.uintptr_t) {
}

//export handle_output_leave
func handle_output_leave(handle C.uintptr_t, output C.uintptr_t) {
}

//export handle_state
func handle_state(handle C.uintptr_t, state C.wl_array) {
	var windowStateMask WindowStateMask = 0
	var size = state.size
	for _, st := range (*[1 << 4]C.uint32_t)(unsafe.Pointer(state.data))[:size:size] {
		switch st {
		case C.ZWLR_FOREIGN_TOPLEVEL_HANDLE_V1_STATE_MAXIMIZED:
			windowStateMask |= MAXIMIZED
		case C.ZWLR_FOREIGN_TOPLEVEL_HANDLE_V1_STATE_MINIMIZED:
			windowStateMask |= MINIMIZED
		case C.ZWLR_FOREIGN_TOPLEVEL_HANDLE_V1_STATE_ACTIVATED:
			windowStateMask |= ACTIVATED
		case C.ZWLR_FOREIGN_TOPLEVEL_HANDLE_V1_STATE_FULLSCREEN:
			windowStateMask |= FULLSCREEN
		}

	}
	windowUpdates <- windowUpdate{wId: uint64(handle), state: windowStateMask + 1}
}

//export handle_done
func handle_done(handle C.uintptr_t) {}

//export handle_parent
func handle_parent(handle C.uintptr_t, parent C.uintptr_t) {}

//export handle_closed
func handle_closed(handle C.uintptr_t) {
	removals <- uint64(handle)
}

func setupAndRunAsWaylandClient() {
	C.initManager()
	for {
		if C.wl_display_dispatch(C.wl_display) == -1 {
			break
		}
	}
}
