#include "_cgo_export.h"
#include <stdio.h>
#include <string.h>
#include <unistd.h>

#include <wayland-client-core.h>
#include "wlr-foreign-toplevel-management-unstable-v1-client-protocol.h"

struct wl_display *wl_display;
struct wl_seat *wl_seat;

// Requests
void close_toplevel(uintptr_t handle) {
	zwlr_foreign_toplevel_handle_v1_close((toplevel_handle)handle);
	wl_display_flush(wl_display);
}

void activate_toplevel(uintptr_t handle) {
	zwlr_foreign_toplevel_handle_v1_activate((toplevel_handle)handle, wl_seat);
	wl_display_flush(wl_display);
}

void hide_toplevel(uintptr_t handle) {
	zwlr_foreign_toplevel_handle_v1_set_minimized((toplevel_handle)handle);
	wl_display_flush(wl_display);
}

void show_toplevel(uintptr_t handle) {
	zwlr_foreign_toplevel_handle_v1_unset_minimized((toplevel_handle)handle);
	wl_display_flush(wl_display);
}

// Events
void tl_handle_title(void *data, struct zwlr_foreign_toplevel_handle_v1 *handle, const char *title) {
	handle_title((uintptr_t)handle, (char*)title);
}

void tl_handle_app_id(void *data, struct zwlr_foreign_toplevel_handle_v1 *handle, const char *app_id) {
	handle_app_id((uintptr_t)handle, (char*)app_id);
}

void tl_handle_output_enter(void *data, struct zwlr_foreign_toplevel_handle_v1 *handle, struct wl_output *output) {
	handle_output_enter((uintptr_t)handle, (uintptr_t)output);
}

void tl_handle_output_leave(void *data, struct zwlr_foreign_toplevel_handle_v1 *handle, struct wl_output *output) {
	handle_output_leave((uintptr_t)handle, (uintptr_t)output);
}

void tl_handle_state(void *data, struct zwlr_foreign_toplevel_handle_v1 *handle, struct wl_array *state) {
	handle_state((uintptr_t)handle, state);
}

void tl_handle_done(void *data, struct zwlr_foreign_toplevel_handle_v1 *handle) {
	handle_done((uintptr_t)handle);
}

void tl_handle_parent(void *data, struct zwlr_foreign_toplevel_handle_v1 *handle, struct zwlr_foreign_toplevel_handle_v1 *parent) {
}

void tl_handle_closed(void *data, struct zwlr_foreign_toplevel_handle_v1 *handle) {
	handle_closed((uintptr_t)handle);
}



struct zwlr_foreign_toplevel_handle_v1_listener toplevel_handle_impl = {
     .title = tl_handle_title,
     .app_id = tl_handle_app_id,
     .output_enter = tl_handle_output_enter,
     .output_leave = tl_handle_output_leave,
     .state = tl_handle_state,
     .done = tl_handle_done,
     .closed = tl_handle_closed,
     .parent = tl_handle_parent,
};


void handle_toplevel(
		void *data,
		struct zwlr_foreign_toplevel_manager_v1 *manager,
		struct zwlr_foreign_toplevel_handle_v1 *tl_handle) {
	zwlr_foreign_toplevel_handle_v1_add_listener(tl_handle, &toplevel_handle_impl, NULL);

}

void handle_finished(
		void *data, 
		struct zwlr_foreign_toplevel_manager_v1 *manager) {
}

struct zwlr_foreign_toplevel_manager_v1_listener toplevel_listener = {
    .toplevel = handle_toplevel,
    .finished = handle_finished,
};

void registerManager(struct wl_registry* registry, uint32_t name, uint32_t version) {
	struct zwlr_foreign_toplevel_manager_v1 *manager = (struct zwlr_foreign_toplevel_manager_v1 *) wl_registry_bind(registry, name, &zwlr_foreign_toplevel_manager_v1_interface, version);
	int i = zwlr_foreign_toplevel_manager_v1_add_listener(manager, &toplevel_listener, NULL);
}

void register_seat(struct wl_registry *registry, uint32_t name, uint32_t version) {
   wl_seat = (struct wl_seat*) wl_registry_bind(registry, name, &wl_seat_interface, version);
}
 
void handle_global(void *data, struct wl_registry *registry, uint32_t name, const char *interface, uint32_t version) {
  if (strcmp(interface, zwlr_foreign_toplevel_manager_v1_interface.name) == 0) {
   	registerManager(registry, name, version); 
  } else if (strcmp(interface, wl_seat_interface.name) == 0) {
	register_seat(registry, name, version);
  }
}

void handle_global_remove(void *data, struct wl_registry *registry, uint32_t name) {
}

struct wl_registry_listener registry_listener_impl = {
	.global = handle_global,
	.global_remove = handle_global_remove
};

void initManager() {
	wl_display = wl_display_connect(NULL);
	struct wl_registry *registry = wl_display_get_registry(wl_display);
 	wl_registry_add_listener(registry, &registry_listener_impl, NULL);
  	wl_display_roundtrip(wl_display);
}



