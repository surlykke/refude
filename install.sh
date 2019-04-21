#!/usr/bin/env bash
thisdir=$(dirname $(realpath $0))
cd $thisdir
rm -rf ${thisdir}/dist/*

[[ -n "$PREFIX" ]] || PREFIX=$HOME/.local
REFUDEDIR=${PREFIX}/share/refude
mkdir -p ${REFUDEDIR}

./build.sh

echo done building

for appdir in panel appchooser ; do
	rm -rf $REFUDEDIR/$appdir
	cp -R dist/$appdir $REFUDEDIR/$appdir
done

for app in panel/refudePanel panel/refudeDo appchooser/refudeAppChooser;  do
	ln -sf $REFUDEDIR/$app ${PREFIX}/bin
done

for desktopfile in panel/refudePanel.desktop; do
    ln -sf $REFUDEDIR/$desktopfile ${PREFIX}/share/applications
done
