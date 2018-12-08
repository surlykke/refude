#!/usr/bin/env bash
thisdir=$(dirname $(realpath $0))
cd $thisdir
rm -rf ${thisdir}/dist/*

[[ -n "$PREFIX" ]] || PREFIX=$HOME/.local
REFUDEDIR=${PREFIX}/share/refude
mkdir -p ${REFUDEDIR}

./build.sh

for appdir in panel appchooser notifications ; do
	cp -R dist/$appdir ${PREFIX}/share/refude/$appdir
done

for app in panel/refudePanel panel/refudeDo appchooser/refudeAppChooser notifications/refudeNotifications;  do
	ln -sf ${PREFIX}/share/refude/$app ${PREFIX}/bin
done

for desktopfile in panel/refudePanel.desktop notifications/refudeNotifications.desktop; do
    ln -sf ${PREFIX}/share/refude/$desktopfile ${PREFIX}/share/applications
done
