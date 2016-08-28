#!/usr/bin/env bash

function exitIfNotRunning {
	echo -n .
	curl --unix $XDG_RUNTIME_DIR/org.restfulipc.refude.desktop http://localhost >/dev/null 2>/dev/null
	if [[ "x7" == "x$?" ]]; then
		echo	
		echo "refude is stopped"
		exit 0
	fi
}

curl -X POST --unix $XDG_RUNTIME_DIR/org.restfulipc.refude.desktop http://localhost/quit >/dev/null 2>/dev/null

for i in {1..50}; do
	exitIfNotRunning
	sleep 0.1
done

echo "Stopping refude failed"
