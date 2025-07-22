#!/usr/bin/env bash
# Very simple script to run refude desktop service
# Call from inside a DE-session, as (at least) PATH and DISPLAY variables
# are needed by the services.
#

killall refude-server
# Wait for port to 7938 be released
for i in {1..10}; do
	if ! netstat -ltnp 2>/dev/null | grep 7938; then 
		break
	fi
	sleep 0.5
done

# gtk4-layer-shell lib must be loaded before any wayland libs
GTK_LAYER_SHELL_LIB=$(ldconfig -p | grep 'gtk4-layer-shell.so$' | sed 's/.*=>\s*//')
LOGFILE=${XDG_RUNTIME_DIR:-/tmp}/RefudeServices.log
LD_PRELOAD=$GTK_LAYER_SHELL_LIB nohup refude-server $REFUDE_SWITCHES >$LOGFILE 2>$LOGFILE &

