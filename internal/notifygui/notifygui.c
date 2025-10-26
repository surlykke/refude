// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
#include <gtk/gtk.h>
#include <gtk4-layer-shell.h>
#include <stdio.h>

char *css = ".subjectLabel {font-size: 20px;} .bodyLabel {font-size: 16px;}";

GtkApplication *application;
GtkWidget *win, *list;

extern void GuiReady();

static void setup(GtkApplication *app, gpointer user_data)
{
    printf("Ind i setup\n");
    GdkDisplay *display = gdk_display_get_default();
    GtkCssProvider *cssProvider = gtk_css_provider_new();
    gtk_css_provider_load_from_string(cssProvider, css);
    gtk_style_context_add_provider_for_display(display, GTK_STYLE_PROVIDER(cssProvider), 1);
    win = gtk_application_window_new(app);
    gtk_window_set_default_size(GTK_WINDOW(win), 1, 1);

    gtk_layer_init_for_window(GTK_WINDOW(win));

    char *default_monitor = getenv("DEFAULT_MONITOR");
    if (default_monitor == NULL)
    {
        default_monitor = "eDP-1";
    }
    GListModel *monitors = gdk_display_get_monitors(display);
    for (int i = 0;; i++)
    {
        GdkMonitor *m = g_list_model_get_item(monitors, i);
        if (m != NULL)
        {
            if (strcmp(default_monitor, gdk_monitor_get_connector(m)) == 0)
            {
                gtk_layer_set_monitor(GTK_WINDOW(win), m);
                printf("monitor set to %s\n", gdk_monitor_get_connector(m));
                break;
            }
        }
        else
        {
            break;
        }
    }

    gtk_layer_set_layer(GTK_WINDOW(win), GTK_LAYER_SHELL_LAYER_OVERLAY);
    gtk_layer_set_anchor(GTK_WINDOW(win), GTK_LAYER_SHELL_EDGE_TOP, true);
    gtk_layer_set_anchor(GTK_WINDOW(win), GTK_LAYER_SHELL_EDGE_RIGHT, true);
    gtk_layer_set_margin(GTK_WINDOW(win), GTK_LAYER_SHELL_EDGE_TOP, 5);
    gtk_layer_set_margin(GTK_WINDOW(win), GTK_LAYER_SHELL_EDGE_RIGHT, 5);

    list = gtk_box_new(GTK_ORIENTATION_VERTICAL, 8);
    gtk_window_set_child(GTK_WINDOW(win), list);
    GuiReady();
    printf("Ud af setup\n");
}

void run()
{
    printf("Ind i run\n");
    application = gtk_application_new("org.refude.notify", G_APPLICATION_DEFAULT_FLAGS);
    g_signal_connect(application, "activate", G_CALLBACK(setup), NULL);

    int status;

    status = g_application_run(G_APPLICATION(application), 0, NULL);
    g_object_unref(application);
    printf("Ud af run\n");
}

struct flash_data
{
    char **notifications;
    int num;
};

int updateInMainLoop(void *dataV)
{
    struct flash_data *data = (struct flash_data *)(dataV);

    // Clear

    GtkWidget *notification = gtk_widget_get_first_child(list);
    while (notification != NULL)
    {
        gtk_box_remove(GTK_BOX(list), notification);
        notification = gtk_widget_get_first_child(list);
    }

    for (int i = 0; i < data->num; i++)
    {
        char *subject = data->notifications[3 * i];
        char *body = data->notifications[3 * i + 1];
        char *iconPath = data->notifications[3 * i + 2];

        GtkWidget *hbox, *iconImage, *vbox, *subjectLabel, *bodyLabel;
        hbox = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 5);
        gtk_widget_add_css_class(hbox, "notification");
        gtk_box_append(GTK_BOX(list), hbox);

        iconImage = gtk_image_new();
        gtk_widget_set_margin_start(iconImage, 8);
        if (iconPath != NULL)
        {
            gtk_image_set_from_file(GTK_IMAGE(iconImage), iconPath);
        }
        gtk_image_set_pixel_size(GTK_IMAGE(iconImage), 48);
        gtk_box_append(GTK_BOX(hbox), iconImage);

        vbox = gtk_box_new(GTK_ORIENTATION_VERTICAL, 12);
        gtk_widget_set_margin_start(vbox, 12);
        gtk_widget_set_margin_top(vbox, 8);
        gtk_widget_set_margin_end(vbox, 12);
        gtk_widget_set_margin_bottom(vbox, 12);
        gtk_box_append(GTK_BOX(hbox), vbox);

        subjectLabel = gtk_label_new(subject);
        gtk_widget_add_css_class(subjectLabel, "subjectLabel");
        gtk_label_set_xalign(GTK_LABEL(subjectLabel), 0);
        gtk_box_append(GTK_BOX(vbox), subjectLabel);

        bodyLabel = gtk_label_new(body);
        gtk_widget_add_css_class(bodyLabel, "bodyLabel");
        gtk_label_set_wrap(GTK_LABEL(bodyLabel), true);
        gtk_box_append(GTK_BOX(vbox), bodyLabel);
        gtk_label_set_xalign(GTK_LABEL(bodyLabel), 0);
    }
    gtk_widget_set_visible(win, data->num > 0);

    return 0;
}

void update(char **notifications, int number)
{
    struct flash_data *data = g_rc_box_new(struct flash_data);
    data->notifications = notifications;
    data->num = number;
    g_idle_add(updateInMainLoop, data);
}
