#!/usr/bin/env bash

curl -X POST --unix $XDG_RUNTIME_DIR/org.restfulipc.refude.desktop http://localhost/quit >/dev/null 2>/dev/null
