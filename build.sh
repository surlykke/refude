#!/usr/bin/env bash

[[ -z "$1" ]] || [[ "-d" == "$1" ]] || { echo "Usage: build.sh [-d]" > /dev/stderr; exit 1; }

cd $(dirname $0)

for appdir in appchooser panel; do
    mkdir -p dist/$appdir
    cp $appdir/static/* dist/$appdir
done

if [[ "-d" == "$1" ]]; then
    webpack -d
else
    webpack -p
fi
