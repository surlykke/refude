#!/bin/bash
cd `dirname $0`

BIN_DIR=${HOME}/.local/bin
GOPATH=${GOPATH:-${HOME}/go}
GO_BIN_DIR=${GOPATH}/bin
BASH_COMPLETION_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/bash/completions
FISH_COMPLETION_DIR=${XDG_CONFIG_HOME:-${HOME}/.config}/fish/completions
mkdir -p $BIN_DIR $BASH_COMPLETION_DIR $FISH_COMPLETION_DIR

for dir in refuc RefudeXdgOpen RefudeDesktopService; do
    if [[ -d $dir ]]; then
        (cd $dir && echo "Building $dir" && go install) || exit 1
    fi
done

cp scripts/bin/runRefude.sh $BIN_DIR
cp scripts/completions/bash/* ${BASH_COMPLETION_DIR}
cp scripts/completions/fish/* ${FISH_COMPLETION_DIR}
ln -f $GO_BIN_DIR/RefudeXdgOpen $BIN_DIR/xdg-open
