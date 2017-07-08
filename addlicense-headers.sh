#!/bin/bash
cd $(dirname $0)
DIRS="appchooser common connman do panel power"

addHeader () {
	cat ./license-header.txt $1 | sponge $1
}
export -f addHeader

find lib Refude* -name '*.go' -not -exec grep -q GPL2 {} \; -exec bash -c 'addHeader "{}"' \; -print

