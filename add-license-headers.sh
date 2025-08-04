#!/bin/bash
cd $(dirname $0)

HEADER=$(
cat<<-EOF
// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
EOF
)


for path in $(find . \( -name '*.go' -o -name '*.js' -o -name '*.c' -o -name '*.h' \) \
		-not -name NamedColors.go \
		-not -name scheme-handlers.go \
		-not -name wlr-foreign-toplevel-management-unstable-v1-client-protocol.h \
		-not -name sse.js \
		-not -name htmx.min.js \
		-not -exec grep -q GPL2 {} \; -print); do
	if [[ "--dry-run" == "$1" ]]; then
		echo $path
	else 
		tmp=`mktemp`
		echo "$HEADER" > $tmp
		cat $path >> $tmp
		mv $tmp $path
	fi
done
