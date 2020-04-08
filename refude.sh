#!/usr/bin/env bash
if ! curl -D - --unix $XDG_RUNTIME_DIR/org.refude.panel.do http://localhost/$1 ; then
    killall refude
    unlink $XDG_RUNTIME_DIR/org.refude.panel.do
    nohup refude >/dev/null 2>/dev/null &
else 
    echo curlie
fi
