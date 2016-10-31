#!/usr/bin/env bash

THISDIR=$(dirname $(realpath $0))
cd `realpath "$THISDIR/.."`

nohup npm start >/tmp/refude.log 2>&1 &
