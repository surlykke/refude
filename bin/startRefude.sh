#!/usr/bin/env bash

THISDIR=$(dirname $(realpath $0))
cd `realpath "$THISDIR/.."`

function exitIfRunning {
	if curl --unix $XDG_RUNTIME_DIR/org.restfulipc.refude.desktop http://localhost >/dev/null 2>/dev/null; then
		echo	
		echo "refude is started."
		exit 0
	fi
}

exitIfRunning;

nohup npm start >/dev/null 2>&1 &

for i in {1..50}; do
	sleep 0.1
	echo -n .
	exitIfRunning;
done

echo starting refude failed
