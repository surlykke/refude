#!/usr/bin/env bash

[[ -z "$1" ]] || [[ "-d" == "$1" ]] || { echo "Usage: build.sh [-d]" > /dev/stderr; exit 1; }

cd $(dirname $0)


mkdir -p dist/panel
cp panel/static/* dist/panel

if [[ "-d" == "$1" ]]; then
    ./node_modules/.bin/webpack -d
else
    ./node_modules/.bin/webpack -p
fi
