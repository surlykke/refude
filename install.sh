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

go install 
cd ${thisdir}/tools/refuc 
go install 
cd ${thisdir}/notifyclient
go install
cd ${thisdir}

cp README.md $ASSETS_DIR
cp -R ./refudeicons/* $HICOLOR_ICON_DIR 
cp resources/bin/* $GOBIN
cp resources/completions/bash/* ${BASH_COMPLETION_DIR}
cp resources/completions/fish/* ${FISH_COMPLETION_DIR}
