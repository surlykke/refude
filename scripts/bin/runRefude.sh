#!/usr/bin/env bash
# Very simple script to run refude services
# Call from inside a DE-session, as (at least) PATH and DISPLAY variables
# are needed by the services.
#

REFUDESERVICES="RefudeDesktopService RefudeIconService RefudePowerService RefudeWmService RefudeStatusNotifierService RefudeNotificationService RefudeActionService RefudeProxy"

if [[ "--restart" == "$1" ]]; then
    for app in $REFUDESERVICES; do
        killall $app
    done
fi

GOPATH="${GOPATH:-$HOME/go}"

# Run refudeservices.
for app in $REFUDESERVICES; do
	nohup $app >/tmp/${app}.log 2>/tmp/${app}.log &
done

