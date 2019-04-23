#!/usr/bin/env bash
# Very simple script to run refude desktop service
# Call from inside a DE-session, as (at least) PATH and DISPLAY variables
# are needed by the services.
#

if [[ "--restart" == "$1" ]]; then
	kill `pgrep -f RefudeDesktopService`
fi

LOGFILE=/tmp/RefudeDesktopService_`date +%s`.log
nohup RefudeDesktopService >$LOGFILE 2>$LOGFILE &

