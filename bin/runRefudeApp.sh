#!/bin/sh
echo "Trying to start $1"
curl -X POST --unix-socket /run/user/1000/org.restfulipc.refude.desktop http://localhost/$1;
if [[ "x7" == "x$$?" ]]; then
	echo "Refude seemingly not running.."
fi

