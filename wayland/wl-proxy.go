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
void hide_toplevel(uintptr_t);
void show_toplevel(uintptr_t);
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/surlykke/RefudeServices/lib/resourcerepo"
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
	var ww = getCopy(uint64(handle))
	ww.Title = C.GoString(c_title)
	resourcerepo.Put(ww)
}

//export handle_app_id
func handle_app_id(handle C.uintptr_t, c_app_id *C.char) {
	var ww = getCopy(uint64(handle))
	ww.AppId = C.GoString(c_app_id)
	ww.Comment = ww.AppId
	resourcerepo.Put(ww)
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

	var ww = getCopy(uint64(handle))
	ww.State = windowStateMask
	resourcerepo.Put(ww)

}

//export handle_done
func handle_done(handle C.uintptr_t) {}

//export handle_parent
func handle_parent(handle C.uintptr_t, parent C.uintptr_t) {}

//export handle_closed
func handle_closed(handle C.uintptr_t) {
	resourcerepo.Remove(fmt.Sprintf("/window/%d", handle))
}

func setupAndRunAsWaylandClient() {
	C.initManager()
	for {
		if C.wl_display_dispatch(C.wl_display) == -1 {
			break
		}
	}
}
