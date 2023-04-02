package wayland

/*
#cgo CFLAGS: -I. -I${SRCDIR}
#cgo LDFLAGS: -lwayland-client

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
void set_toplevel_rectangle(uintptr_t handle, int32_t x, int32_t y, int32_t width, int32_t height);
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

func setRectangle(wId uint64, x uint32, y uint32, w uint32, h uint32) {
	C.set_toplevel_rectangle(C.uintptr_t(wId), C.int32_t(x), C.int32_t(y), C.int32_t(w), C.int32_t(h))
}

//export handle_title
func handle_title(handle C.uintptr_t, c_title *C.char) {
	WM.handle_title(uint64(handle), C.GoString(c_title))
}

//export handle_app_id
func handle_app_id(handle C.uintptr_t, c_app_id *C.char) {
	WM.handle_app_id(uint64(handle), C.GoString(c_app_id))
}

//export handle_output_enter
func handle_output_enter(handle C.uintptr_t, output C.uintptr_t) {
	WM.handle_output_enter(uint64(handle), uint64(output))
}

//export handle_output_leave
func handle_output_leave(handle C.uintptr_t, output C.uintptr_t) {
	WM.handle_output_leave(uint64(handle), uint64(output))
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

	WM.handle_state(uint64(handle), windowStateMask) 
}

//export handle_done
func handle_done(handle C.uintptr_t) {
	WM.handle_done(uint64(handle))
}

//export handle_parent
func handle_parent(handle C.uintptr_t, parent C.uintptr_t) {
	WM.handle_parent(uint64(handle), uint64(parent))
}

//export handle_closed
func handle_closed(handle C.uintptr_t) {
	WM.handle_closed(uint64(handle))
}

func setupAndRunAsWaylandClient() {
	C.initManager()
	for {
		if C.wl_display_dispatch(C.wl_display) == -1 {
			break
		}
	}
}
