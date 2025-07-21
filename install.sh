#!/bin/bash
cd `dirname $0`
thisdir=$(pwd)
set -x
set -e 

export GOBIN=${HOME}/.local/bin
BASH_COMPLETION_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/bash/completions
FISH_COMPLETION_DIR=${XDG_CONFIG_HOME:-${HOME}/.config}/fish/completions
HICOLOR_ICON_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/icons/hicolor
DESKTOP_FILE_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/applications
ASSETS_DIR=${XDG_DATA_HOME:-${HOME}/.local/share}/RefudeServices
mkdir -p $GOBIN $BASH_COMPLETION_DIR $FISH_COMPLETION_DIR $HICOLOR_ICON_DIR $ASSETS_DIR

cd ${thisdir}/server
go build -o  refude-server && mv ./refude-server $GOBIN # Annoyingly, go install does not allow specifying executable name
cp ./runRefude.sh $GOBIN
cp -R ./refudeicons/* $HICOLOR_ICON_DIR 

cd ${thisdir}/refuc
go install
cp ./completions/bash/* ${BASH_COMPLETION_DIR}
cp ./completions/fish/* ${FISH_COMPLETION_DIR}

