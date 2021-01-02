#!/bin/bash
cd `dirname $0`

GOBIN=${HOME}/.local/bin
BASH_COMPLETION_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/bash/completions
FISH_COMPLETION_DIR=${XDG_CONFIG_HOME:-${HOME}/.config}/fish/completions
HICOLOR_ICON_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/icons/hicolor
ASSETS_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/RefudeServices
mkdir -p $GOBIN $BASH_COMPLETION_DIR $FISH_COMPLETION_DIR $HICOLOR_ICON_DIR $ASSETS_DIR


go install || exit 1
(cd refuc && go install ) || exit 1
cp README.md $ASSETS_DIR
cp -R ./icons/* $HICOLOR_ICON_DIR 
cp scripts/bin/* $GOBIN
cp scripts/completions/bash/* ${BASH_COMPLETION_DIR}
cp scripts/completions/fish/* ${FISH_COMPLETION_DIR}
