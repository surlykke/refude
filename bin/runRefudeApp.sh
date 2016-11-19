#!/usr/bin/env bash

THISDIR=$(dirname $(realpath $0))
APPDIR=`realpath "$THISDIR/../$1"`
[[ -d $APPDIR ]] || { echo "Directory $APPDIR not found"; exit 1; }
exec electron $APPDIR

