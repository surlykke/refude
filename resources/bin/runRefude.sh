#!/usr/bin/env bash
# Very simple script to run refude desktop service
# Call from inside a DE-session, as (at least) PATH and DISPLAY variables
# are needed by the services.
#

if [[ "--restart" == "$1" ]]; then
	pid=$(pgrep -f RefudeServices)
	[[ -n "$pid" ]] && kill $pid
fi

LOGFILE=${XDG_RUNTIME_DIR:-/tmp}/RefudeServices.log
nohup RefudeServices >$LOGFILE 2>$LOGFILE &

