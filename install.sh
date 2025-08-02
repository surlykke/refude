#!/bin/bash
cd `dirname $0`
set -x
set -e 

rm ./refuc ./refude-nm ./refude-server 2>/dev/null || :

export GOBIN=${GOBIN:-${HOME}/.local/bin}
BASH_COMPLETION_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/bash/completions
FISH_COMPLETION_DIR=${XDG_CONFIG_HOME:-${HOME}/.config}/fish/completions
HICOLOR_ICON_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/icons/hicolor
DESKTOP_FILE_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/applications
ASSETS_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/refude
mkdir -p $GOBIN $BASH_COMPLETION_DIR $FISH_COMPLETION_DIR $HICOLOR_ICON_DIR $ASSETS_DIR

go install ./cmd/refude-server
cp ./cmd/refude-server/runRefude.sh $GOBIN
cp -R ./internal/refudeicons/* $HICOLOR_ICON_DIR 

go install ./cmd/refuc
cp ./cmd/refuc/completions/bash/* ${BASH_COMPLETION_DIR}
cp ./cmd/refuc/completions/fish/* ${FISH_COMPLETION_DIR}

go install ./cmd/refude-nm 
sed "s@GOBIN@${GOBIN}@g" ./cmd/refude-nm/org.refude.native_messaging.json > ${ASSETS_DIR}/org.refude.native_messaging.json

