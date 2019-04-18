#!/bin/bash
cd $(dirname $0)

addHeader () {
	cat ./license-header.txt $1 | sponge $1
}
export -f addHeader

find  -name '*.go' -not -exec grep -q GPL2 {} \; -exec bash -c 'addHeader "{}"' \; -print

