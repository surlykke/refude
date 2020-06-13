#!/usr/bin/env bash
thisdir=$(dirname $(realpath $0))
cd $thisdir
rm -rf ${thisdir}/refude-linux-x64

[[ -n "$PREFIX" ]] || PREFIX=$HOME/.local
REFUDEDIR=${PREFIX}/share/refude
mkdir -p ${REFUDEDIR}

./node_modules/electron-packager/bin/electron-packager.js . --asar || exit 1
cp ./refude.sh ./refude.desktop refude-linux-x64

echo Done building
echo Installing under $PREFIX

cp -R refude-linux-x64/* $REFUDEDIR
ln -sf $REFUDEDIR/refude ${PREFIX}/bin
ln -sf $REFUDEDIR/refude.sh ${PREFIX}/bin
ln -sf $REFUDEDIR/refude.desktop ${PREFIX}/share/applications
