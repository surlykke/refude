#include "gtk4-layer-shell.h"
#include <gtk/gtk.h>
// go:embed gui.css

char *css = "#subjectLabel {font-size: 20px;} #bodyLabel {font-size: 16px; }";

GtkApplication *application;
GtkWidget *win, *revealer, *hbox, *iconImage, *vbox, *subjectLabel, *bodyLabel;

static void setup(GtkApplication *app, gpointer user_data) {

  GdkDisplay *display = gdk_display_get_default();
  GtkCssProvider *cssProvider = gtk_css_provider_new();
  gtk_css_provider_load_from_string(cssProvider, css);
  gtk_style_context_add_provider_for_display(
      display, GTK_STYLE_PROVIDER(cssProvider), 1);
  win = gtk_application_window_new(app);

  GdkSurface *surface = gtk_native_get_surface(GTK_NATIVE(win));
  printf("surface: %p\n", surface);	

  gtk_layer_init_for_window(GTK_WINDOW(win));
  char *default_monitor = getenv("DEFAULT_MONITOR");
  if (default_monitor != NULL) {
  	GListModel *monitors = gdk_display_get_monitors(display);
    for (int i = 0;; i++) {
      GdkMonitor *m = g_list_model_get_item(monitors, i);
      if (m != NULL) {
        if (strcmp(default_monitor, gdk_monitor_get_connector(m)) == 0) {
          gtk_layer_set_monitor(GTK_WINDOW(win), m);
          printf("monitor set to %s\n", gdk_monitor_get_connector(m));
          break;
        }
      } else {
        break;
      }
    }
  }

  gtk_layer_set_layer(GTK_WINDOW(win), GTK_LAYER_SHELL_LAYER_TOP);
  gtk_layer_set_anchor(GTK_WINDOW(win), GTK_LAYER_SHELL_EDGE_RIGHT, true);
  gtk_layer_set_anchor(GTK_WINDOW(win), GTK_LAYER_SHELL_EDGE_BOTTOM, true);
  gtk_layer_set_margin(GTK_WINDOW(win), GTK_LAYER_SHELL_EDGE_RIGHT, 5);
  gtk_layer_set_margin(GTK_WINDOW(win), GTK_LAYER_SHELL_EDGE_BOTTOM, 5);

  revealer = gtk_revealer_new();
  gtk_window_set_child(GTK_WINDOW(win), revealer);
  gtk_revealer_set_transition_type(GTK_REVEALER(revealer),
                                   GTK_REVEALER_TRANSITION_TYPE_SLIDE_LEFT);
  gtk_revealer_set_transition_duration(GTK_REVEALER(revealer), 200);
  gtk_revealer_set_reveal_child(GTK_REVEALER(revealer), false);

  hbox = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 5);
  gtk_revealer_set_child(GTK_REVEALER(revealer), hbox);

  iconImage = gtk_image_new();
  gtk_widget_set_margin_start(iconImage, 8);
  gtk_image_set_from_file(GTK_IMAGE(iconImage),
                          "/usr/share/icons/hicolor/64x64/apps/firefox.png");
  gtk_image_set_pixel_size(GTK_IMAGE(iconImage), 48);
  gtk_box_append(GTK_BOX(hbox), iconImage);

  vbox = gtk_box_new(GTK_ORIENTATION_VERTICAL, 12);
  gtk_widget_set_margin_start(vbox, 12);
  gtk_widget_set_margin_top(vbox, 8);
  gtk_widget_set_margin_end(vbox, 12);
  gtk_widget_set_margin_bottom(vbox, 12);

  gtk_box_append(GTK_BOX(hbox), vbox);
  subjectLabel = gtk_label_new("");
  gtk_widget_set_name(subjectLabel, "subjectLabel");
  gtk_label_set_xalign(GTK_LABEL(subjectLabel), 0);
  gtk_box_append(GTK_BOX(vbox), subjectLabel);
  bodyLabel = gtk_label_new("");
  gtk_label_set_wrap(GTK_LABEL(bodyLabel), true);
  gtk_box_append(GTK_BOX(vbox), bodyLabel);
  gtk_label_set_xalign(GTK_LABEL(bodyLabel), 0);
  gtk_widget_set_name(bodyLabel, "bodyLabel");
}

void run() {
  application =
      gtk_application_new("org.refude.notify", G_APPLICATION_DEFAULT_FLAGS);
  g_signal_connect(application, "activate", G_CALLBACK(setup), NULL);

  int status;

  status = g_application_run(G_APPLICATION(application), 0, NULL);
  g_object_unref(application);
}

struct flash_data {
  bool show;
  char *subject;
  char *body;
  char *iconfilePath;
};

int updateInMainLoop(void *dataV) {
  struct flash_data *data = (struct flash_data *)(dataV);
  if (!data->show) {
    gtk_revealer_set_reveal_child(GTK_REVEALER(revealer), false);
  } else {
    gtk_label_set_text(GTK_LABEL(subjectLabel), data->subject);
    gtk_label_set_text(GTK_LABEL(bodyLabel), data->body);
	if (strlen(data->body) > 30) {
		gtk_label_set_width_chars(GTK_LABEL(bodyLabel), 30);
	} else {
		gtk_label_set_width_chars(GTK_LABEL(bodyLabel), -1);
 	}
	printf("c: iconFilePath %s\n", data->iconfilePath);
    if (data->iconfilePath != NULL) {
      gtk_image_set_from_file(GTK_IMAGE(iconImage), data->iconfilePath);
	  gtk_widget_set_visible(iconImage, true);
    } else {
      gtk_image_clear(GTK_IMAGE(iconImage));
	  gtk_widget_set_visible(iconImage, false);
    }
    gtk_window_set_default_size(GTK_WINDOW(win), 1, 1);
    gtk_widget_set_visible(win, true);
    gtk_revealer_set_reveal_child(GTK_REVEALER(revealer), true);
  }
  return 0;
}
// Once the unrevealing complete, clean up...
void update(int show, char *subject, char *body, char *iconFile) {
  struct flash_data *data = g_rc_box_new(struct flash_data);
  data->show = show;
  data->subject = subject;
  data->body = body;
  data->iconfilePath = iconFile;
  g_idle_add(updateInMainLoop, data);
}

int hideInMainLoop(void *data) {
  if (!gtk_revealer_get_reveal_child(GTK_REVEALER(revealer))) {
    gtk_widget_set_visible(win, false);
  }
  return 0;
}

void hide() { g_idle_add(hideInMainLoop, NULL); }
