#!/bin/bash
cd $(dirname $0)

for path in $(find . \( -name '*.go' -o -name '*.js' -o -name '*.c' -o -name '*.h' \) -not -name NamedColors.go -not -name scheme-handlers.go -not -exec grep -q GPL2 {} \; -print); do
	tmp=`mktemp`

cat<< EOF > $tmp
// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
EOF

	cat $path >> $tmp
	mv $tmp $path
done
