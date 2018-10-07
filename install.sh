#!/usr/bin/env bash
thisdir=$(dirname $(realpath $0))
cd $thisdir
rm -rf ${thisdir}/dist/*

[[ -n "$PREFIX" ]] || PREFIX=$HOME/.local
REFUDEDIR=${PREFIX}/share/refude
mkdir -p ${REFUDEDIR}

for appdir in panel appchooser; do
	echo "building $appdir"
	cd $thisdir/$appdir
	gulp || exit 1
	cp -R $thisdir/dist/$appdir ${PREFIX}/share/refude

done

for app in panel/refudePanel panel/refudeDo appchooser/refudeAppChooser; do
	ln -sf ${PREFIX}/share/refude/$app ${PREFIX}/bin
done

ln -sf ${PREFIX}/share/refude/panel/refudePanel.desktop ${PREFIX}/share/applications
