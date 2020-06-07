#!/bin/bash
cd `dirname $0`

BIN_DIR=${HOME}/.local/bin
GOPATH=${GOPATH:-${HOME}/go}
GO_BIN_DIR=${GOPATH}/bin
BASH_COMPLETION_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/bash/completions
FISH_COMPLETION_DIR=${XDG_CONFIG_HOME:-${HOME}/.config}/fish/completions
HICOLOR_ICON_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/icons/hicolor
ASSETS_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/RefudeServices/assets
mkdir -p $BIN_DIR $BASH_COMPLETION_DIR $FISH_COMPLETION_DIR $HICOLOR_ICON_DIR $ASSETS_DIR



for dir in refuc RefudeDesktopService; do
    if [[ -d $dir ]]; then
        (cd $dir && echo "Building $dir" && go install) || exit 1
    fi
done

cp ./assets/* $ASSETS_DIR

cp -R ./icons/* $HICOLOR_ICON_DIR 

cp scripts/bin/* $BIN_DIR
cp scripts/completions/bash/* ${BASH_COMPLETION_DIR}
cp scripts/completions/fish/* ${FISH_COMPLETION_DIR}
