#!/bin/bash
cd $(dirname $0)
DIRS="appchooser common panel"

addHtmlHeader () {
	cat ./license-header-template-html.txt $1 | sponge $1
}
export -f addHtmlHeader

find $DIRS -name '*.html' -not -exec grep -q GPL2 {} \; -exec bash -c 'addHtmlHeader "{}"' \; -print

addJsHeader () {
	cat ./license-header-template-js.txt $1 | sponge $1
}
export -f addJsHeader

find $DIRS -name '*.js' -not -exec grep -q GPL2 {} \; -exec bash -c 'addJsHeader "{}"' \; -print
find $DIRS -name '*.jsx' -not -exec grep -q GPL2 {} \; -exec bash -c 'addJsHeader "{}"' \; -print

