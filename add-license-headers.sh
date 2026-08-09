#!/usr/bin/env bash
cd $(dirname $0)

HEADER=$(
cat<<-EOF
// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
//
EOF
)


for path in $(find . \( -name '*.go' -o -name '*.js' -o -name '*.c' -o -name '*.h' \) \
		-not -name NamedColors.go \
		-not -name scheme-handlers.go \
		-not -name wlr-foreign-toplevel-management-unstable-v1-client-protocol.h \
		-not -name htmx.2.0.10.min.js \
		-not -name alpine.3.15.12.min.js \
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
