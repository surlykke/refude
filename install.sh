#!/usr/bin/env bash
thisdir=$(dirname $(realpath $0))
cd $thisdir
rm -rf ${thisdir}/dist/*

[[ -n "$PREFIX" ]] || PREFIX=$HOME/.local
REFUDEDIR=${PREFIX}/share/refude
mkdir -p ${REFUDEDIR}

./build.sh

echo done building

rm -rf $REFUDEDIR/panel
cp -R dist/panel $REFUDEDIR/panel
ln -sf $REFUDEDIR/panel ${PREFIX}/bin
ln -sf $REFUDEDIR/refudePanel.desktop ${PREFIX}/share/applications
