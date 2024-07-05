#!/usr/bin/env bash
# Very simple script to run refude desktop service
# Call from inside a DE-session, as (at least) PATH and DISPLAY variables
# are needed by the services.
#

killall RefudeServices
# Wait for port 7938 be released
for i in {1..10}; do
	if ! netstat -ltnp 2>/dev/null | grep 7938; then 
		break
	fi
	sleep 0.5
done

LOGFILE=${XDG_RUNTIME_DIR:-/tmp}/RefudeServices.log
nohup RefudeServices $REFUDE_SWITCHES >$LOGFILE 2>$LOGFILE &

